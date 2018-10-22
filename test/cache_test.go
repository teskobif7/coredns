package test

import (
	"testing"

	"github.com/coredns/coredns/plugin/proxy"
	"github.com/coredns/coredns/plugin/test"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

func TestLookupCache(t *testing.T) {
	// Start auth. CoreDNS holding the auth zone.
	name, rm, err := test.TempFile(".", exampleOrg)
	if err != nil {
		t.Fatalf("Failed to create zone: %s", err)
	}
	defer rm()

	corefile := `example.org:0 {
       file ` + name + `
}
`
	i, udp, _, err := CoreDNSServerAndPorts(corefile)
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s", err)
	}
	defer i.Stop()

	// Start caching proxy CoreDNS that we want to test.
	corefile = `example.org:0 {
	proxy . ` + udp + `
	cache 10
}
`
	i, udp, _, err = CoreDNSServerAndPorts(corefile)
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s", err)
	}
	defer i.Stop()

	p := proxy.NewLookup([]string{udp})
	state := request.Request{W: &test.ResponseWriter{}, Req: new(dns.Msg)}

	t.Run("Long TTL", func(t *testing.T) {
		testCase(t, state, p, "example.org.", 2, 10)
	})

	t.Run("Short TTL", func(t *testing.T) {
		testCase(t, state, p, "short.example.org.", 1, 5)
	})

}

func testCase(t *testing.T, state request.Request, p proxy.Proxy, name string, expectAnsLen int, expectTTL uint32) {
	resp, err := p.Lookup(state, name, dns.TypeA)
	if err != nil {
		t.Fatal("Expected to receive reply, but didn't")
	}

	if len(resp.Answer) != expectAnsLen {
		t.Fatalf("Expected %v RR in the answer section, got %v.", expectAnsLen, len(resp.Answer))
	}

	ttl := resp.Answer[0].Header().Ttl
	if ttl != expectTTL {
		t.Errorf("Expected TTL to be %d, got %d", expectTTL, ttl)
	}
}
