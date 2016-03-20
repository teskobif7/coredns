package proxy

// function OTHER middleware might want to use to do lookup in the same
// style as the proxy.

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/miekg/coredns/middleware"
	"github.com/miekg/dns"
)

func New(hosts []string) Proxy {
	p := Proxy{Next: nil, Client: Clients()}

	upstream := &staticUpstream{
		from:         "",
		proxyHeaders: make(http.Header),
		Hosts:        make([]*UpstreamHost, len(hosts)),
		Policy:       &Random{},
		FailTimeout:  10 * time.Second,
		MaxFails:     1,
	}

	for i, host := range hosts {
		uh := &UpstreamHost{
			Name:         host,
			Conns:        0,
			Fails:        0,
			FailTimeout:  upstream.FailTimeout,
			Unhealthy:    false,
			ExtraHeaders: upstream.proxyHeaders,
			CheckDown: func(upstream *staticUpstream) UpstreamHostDownFunc {
				return func(uh *UpstreamHost) bool {
					if uh.Unhealthy {
						return true
					}
					if uh.Fails >= upstream.MaxFails &&
						upstream.MaxFails != 0 {
						return true
					}
					return false
				}
			}(upstream),
			WithoutPathPrefix: upstream.WithoutPathPrefix,
		}
		upstream.Hosts[i] = uh
	}
	p.Upstreams = []Upstream{upstream}
	return p
}

func (p Proxy) Lookup(state middleware.State, name string, tpe uint16) (*dns.Msg, error) {
	req := new(dns.Msg)
	req.SetQuestion(name, tpe)
	// TODO(miek):
	// USE STATE FOR DNSSEC ETCD BUFSIZE BLA BLA
	return p.lookup(state, req)
}

func (p Proxy) lookup(state middleware.State, r *dns.Msg) (*dns.Msg, error) {
	var (
		reply *dns.Msg
		err   error
	)
	for _, upstream := range p.Upstreams {
		// allowed bla bla bla TODO(miek): fix full proxy spec from caddy
		start := time.Now()

		// Since Select() should give us "up" hosts, keep retrying
		// hosts until timeout (or until we get a nil host).
		for time.Now().Sub(start) < tryDuration {
			host := upstream.Select()
			if host == nil {
				return nil, errUnreachable
			}

			atomic.AddInt64(&host.Conns, 1)
			// tls+tcp ?
			if state.Proto() == "tcp" {
				reply, err = middleware.Exchange(p.Client.TCP, r, host.Name)
			} else {
				reply, err = middleware.Exchange(p.Client.UDP, r, host.Name)
			}
			atomic.AddInt64(&host.Conns, -1)

			if err == nil {
				return reply, nil
			}
			timeout := host.FailTimeout
			if timeout == 0 {
				timeout = 10 * time.Second
			}
			atomic.AddInt32(&host.Fails, 1)
			go func(host *UpstreamHost, timeout time.Duration) {
				time.Sleep(timeout)
				atomic.AddInt32(&host.Fails, -1)
			}(host, timeout)
		}
		return nil, errUnreachable
	}
	return nil, errUnreachable
}
