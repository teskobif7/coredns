package test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/coredns/coredns/core/dnsserver"

	"github.com/miekg/dns"
)

// Start 2 tests server, server A will proxy to B, server B is an CH server.
func TestProxyToChaosServer(t *testing.T) {
	t.Parallel()
	corefile := `.:0 {
		chaos CoreDNS-001 miek@miek.nl
	}`

	chaos, udpChaos, _, err := CoreDNSServerAndPorts(corefile)
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s", err)
	}

	defer chaos.Stop()

	corefileProxy := `.:0 {
		forward . ` + udpChaos + `
	}`

	proxy, udp, _, err := CoreDNSServerAndPorts(corefileProxy)
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance")
	}
	defer proxy.Stop()

	chaosTest(t, udpChaos)

	chaosTest(t, udp)
	// chaosTest(t, tcp, "tcp"), commented out because we use the original transport to reach the
	// proxy and we only forward to the udp port.
}

func chaosTest(t *testing.T, server string) {
	m := new(dns.Msg)
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Qclass: dns.ClassCHAOS, Name: "version.bind.", Qtype: dns.TypeTXT}

	r, err := dns.Exchange(m, server)
	if err != nil {
		t.Fatalf("Could not send message: %s", err)
	}
	if r.Rcode != dns.RcodeSuccess || len(r.Answer) == 0 {
		t.Fatalf("Expected successful reply, got %s", dns.RcodeToString[r.Rcode])
	}
	if r.Answer[0].String() != `version.bind.	0	CH	TXT	"CoreDNS-001"` {
		t.Fatalf("Expected version.bind. reply, got %s", r.Answer[0].String())
	}
}

func TestReverseExpansion(t *testing.T) {
	// this test needs a fixed port, because with :0 the expanded reverse zone will listen on different
	// addresses and we can't check which ones...
	corefile := `10.0.0.0/15:5053 {
		whoami
	}`

	server, udp, _, err := CoreDNSServerAndPorts(corefile)
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s", err)
	}

	defer server.Stop()

	m := new(dns.Msg)
	m.SetQuestion("whoami.0.10.in-addr.arpa.", dns.TypeA)

	r, err := dns.Exchange(m, udp)
	if err != nil {
		t.Fatalf("Could not send message: %s", err)
	}
	if r.Rcode != dns.RcodeSuccess {
		t.Errorf("Expected NOERROR, got %d", r.Rcode)
	}
	if len(r.Extra) != 2 {
		t.Errorf("Expected 2 RRs in additional section, got %d", len(r.Extra))
	}

	m.SetQuestion("whoami.1.10.in-addr.arpa.", dns.TypeA)
	r, err = dns.Exchange(m, udp)
	if err != nil {
		t.Fatalf("Could not send message: %s", err)
	}
	if r.Rcode != dns.RcodeSuccess {
		t.Errorf("Expected NOERROR, got %d", r.Rcode)
	}
	if len(r.Extra) != 2 {
		t.Errorf("Expected 2 RRs in additional section, got %d", len(r.Extra))
	}

	// should be refused
	m.SetQuestion("whoami.2.10.in-addr.arpa.", dns.TypeA)
	r, err = dns.Exchange(m, udp)
	if err != nil {
		t.Fatalf("Could not send message: %s", err)
	}
	if r.Rcode != dns.RcodeRefused {
		t.Errorf("Expected REFUSED, got %d", r.Rcode)
	}
	if len(r.Extra) != 0 {
		t.Errorf("Expected 0 RRs in additional section, got %d", len(r.Extra))
	}
}

func TestMultiZoneBlockConfigs(t *testing.T) {
	corefile := `.:40000 .:40001 .:40002 {
		debug
	}`

	server, err := CoreDNSServer(corefile)
	defer server.Stop()

	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s", err)
	}

	// unsafe reflection to read unexported fields "context" and "configs" within context
	ctxVal := reflect.ValueOf(server).Elem().FieldByName("context")
	ctxVal2 := reflect.NewAt(ctxVal.Type(), unsafe.Pointer(ctxVal.UnsafeAddr())).Elem()
	configs := reflect.ValueOf(ctxVal2.Interface()).Elem().FieldByName("configs")
	configs2 := reflect.NewAt(configs.Type(), unsafe.Pointer(configs.UnsafeAddr())).Elem()

	for i := 0; i < 3; i++ {
		v := configs2.Index(i)
		config := v.Interface().(*dnsserver.Config)
		if !config.Debug {
			t.Fatalf("Debug was not set for %s://%s:%s", config.Transport, config.Zone, config.Port)
		}
	}
}
