// +build k8s

package test

import (
	"testing"
	"time"

	"github.com/miekg/coredns/middleware/test"

	"github.com/mholt/caddy"
	"github.com/miekg/dns"
)

// Test data
// TODO: Fix the actual RR values

var dnsTestCases = []test.Case{
	{
		Qname: "svc-1-a.test-1.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("svc-1-a.test-1.svc.cluster.local.      303    IN      A       10.0.0.100"),
		},
	},
	{
		Qname: "bogusservice.test-1.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "svc-1-a.*.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("svc-1-a.test-1.svc.cluster.local.      303    IN      A       10.0.0.100"),
		},
	},
	{
		Qname: "svc-1-a.any.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("svc-1-a.test-1.svc.cluster.local.      303    IN      A       10.0.0.100"),
		},
	},
	{
		Qname: "bogusservice.*.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "bogusservice.any.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "*.test-1.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("svc-1-a.test-1.svc.cluster.local.      303    IN      A       10.0.0.100"),
			test.A("svc-1-b.test-1.svc.cluster.local.      303    IN      A       10.0.0.110"),
			test.A("svc-c.test-1.svc.cluster.local.        303    IN      A       10.0.0.115"),
			test.A("headless-svc.test-1.svc.cluster.local.      303    IN      A       172.17.0.5"),
			test.A("headless-svc.test-1.svc.cluster.local.      303    IN      A       172.17.0.6"),
		},
	},
	{
		Qname: "any.test-1.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("svc-1-a.test-1.svc.cluster.local.      303    IN      A       10.0.0.100"),
			test.A("svc-1-b.test-1.svc.cluster.local.      303    IN      A       10.0.0.110"),
			test.A("svc-c.test-1.svc.cluster.local.        303    IN      A       10.0.0.115"),
			test.A("headless-svc.test-1.svc.cluster.local.      303    IN      A       172.17.0.5"),
			test.A("headless-svc.test-1.svc.cluster.local.      303    IN      A       172.17.0.6"),
		},
	},
	{
		Qname: "any.test-2.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "*.test-2.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "*.*.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("svc-1-a.test-1.svc.cluster.local.      303    IN      A       10.0.0.100"),
			test.A("svc-1-b.test-1.svc.cluster.local.      303    IN      A       10.0.0.110"),
			test.A("svc-c.test-1.svc.cluster.local.        303    IN      A       10.0.0.115"),
			test.A("headless-svc.test-1.svc.cluster.local.      303    IN      A       172.17.0.5"),
			test.A("headless-svc.test-1.svc.cluster.local.      303    IN      A       172.17.0.6"),
		},
	},
	{
		Qname: "headless-svc.test-1.svc.cluster.local.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("headless-svc.test-1.svc.cluster.local.      303    IN      A       172.17.0.5"),
			test.A("headless-svc.test-1.svc.cluster.local.      303    IN      A       172.17.0.6"),
		},
	},
	//TODO: Fix below to all use test.SRV not test.A!
	{
		Qname: "_*._*.svc-1-a.test-1.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("_http._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 80 svc-1-a.test-1.svc.cluster.local."),
			test.SRV("_https._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 443 svc-1-a.test-1.svc.cluster.local."),
		},
	},
	{
		Qname: "_*._*.bogusservice.test-1.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "_*._*.svc-1-a.*.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("_http._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 80 svc-1-a.test-1.svc.cluster.local."),
			test.SRV("_https._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 443 svc-1-a.test-1.svc.cluster.local."),
		},
	},
	{
		Qname: "_*._*.svc-1-a.any.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("_http._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 80 svc-1-a.test-1.svc.cluster.local."),
			test.SRV("_https._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 443 svc-1-a.test-1.svc.cluster.local."),
		},
	},
	{
		Qname: "_*._*.bogusservice.*.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "_*._*.bogusservice.any.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "_c-port._*.*.test-1.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("_c-port._udp.svc-c.test-1.svc.cluster.local.      303    IN    SRV 10 100 1234 svc-c.test-1.svc.cluster.local."),
			test.SRV("_c-port._udp.headless-svc.test-1.svc.cluster.local.      303    IN    SRV 10 100 1234 172-17-0-5.headless-svc.test-1.svc.cluster.local."),
			test.SRV("_c-port._udp.headless-svc.test-1.svc.cluster.local.      303    IN    SRV 10 100 1234 172-17-0-6.headless-svc.test-1.svc.cluster.local."),
		},
	},
	{
		Qname: "_*._tcp.any.test-1.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("_http._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 80 svc-1-a.test-1.svc.cluster.local."),
			test.SRV("_https._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 443 svc-1-a.test-1.svc.cluster.local."),
			test.SRV("_http._tcp.svc-1-b.test-1.svc.cluster.local.      303    IN    SRV 10 100 80 svc-1-b.test-1.svc.cluster.local."),
		},
	},
	{
		Qname: "_*._*.any.test-2.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "_*._*.*.test-2.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "_http._tcp.*.*.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.SRV("_http._tcp.svc-1-a.test-1.svc.cluster.local.      303    IN    SRV 10 100 80 svc-1-a.test-1.svc.cluster.local."),
			test.SRV("_http._tcp.svc-1-b.test-1.svc.cluster.local.      303    IN    SRV 10 100 80 svc-1-b.test-1.svc.cluster.local."),
		},
	},
	{
		Qname: "_*.svc-1-a.test-1.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "_*._not-udp-or-tcp.svc-1-a.test-1.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "svc-1-a.test-1.svc.cluster.local.", Qtype: dns.TypeSRV,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
	{
		Qname: "10-20-0-101.test-1.pod.cluster.local.", Qtype: dns.TypeA,
		Rcode:  dns.RcodeServerFailure,
		Answer: []dns.RR{},
	},
	{
		Qname: "123.0.0.10.in-addr.arpa.", Qtype: dns.TypePTR,
		Rcode:  dns.RcodeSuccess,
		Answer: []dns.RR{},
	},
	{
		Qname: "100.0.0.10.in-addr.arpa.", Qtype: dns.TypePTR,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.PTR("100.0.0.10.in-addr.arpa.      303    IN      PTR       svc-1-a.test-1.svc.cluster.local."),
		},
	},
	{
		Qname: "115.0.0.10.in-addr.arpa.", Qtype: dns.TypePTR,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.PTR("115.0.0.10.in-addr.arpa.      303    IN      PTR       svc-c.test-1.svc.cluster.local."),
		},
	},
}

var dnsTestCasesPodsInsecure = []test.Case{
	{
		Qname: "10-20-0-101.test-1.pod.cluster.local.", Qtype: dns.TypeA,
		Rcode: dns.RcodeSuccess,
		Answer: []dns.RR{
			test.A("10-20-0-101.test-1.pod.cluster.local. 0 IN A    10.20.0.101"),
		},
	},
	{
		Qname: "10-20-0-101.test-X.pod.cluster.local.", Qtype: dns.TypeA,
		Rcode:  dns.RcodeNameError,
		Answer: []dns.RR{},
	},
}

func createTestServer(t *testing.T, corefile string) (*caddy.Instance, string) {
	server, err := CoreDNSServer(corefile)
	if err != nil {
		t.Fatalf("Could not get CoreDNS serving instance: %s", err)
	}

	udp, _ := CoreDNSServerPorts(server, 0)
	if udp == "" {
		t.Fatalf("Could not get UDP listening port")
	}

	return server, udp
}

func doIntegrationTests(t *testing.T, corefile string, testCases []test.Case) {
	server, udp := createTestServer(t, corefile)
	defer server.Stop()

	// Work-around for timing condition that results in no-data being returned in
	// test environment.
	time.Sleep(5 * time.Second)

	for _, tc := range testCases {

		dnsClient := new(dns.Client)
		dnsMessage := new(dns.Msg)

		dnsMessage.SetQuestion(tc.Qname, tc.Qtype)

		res, _, err := dnsClient.Exchange(dnsMessage, udp)
		if err != nil {
			t.Fatalf("Could not send query: %s", err)
		}

		// check the answer
		if res.Rcode != tc.Rcode {
			t.Errorf("Expected rcode %d but got %d for query %s, %d", tc.Rcode, res.Rcode, tc.Qname, tc.Qtype)
		}

		if len(res.Answer) != len(tc.Answer) {
			t.Errorf("Expected %d answers but got %d for query %s, %d", len(tc.Answer), len(res.Answer), tc.Qname, tc.Qtype)
		}

		//TODO: Check the actual RR values
	}
}

func TestKubernetesIntegration(t *testing.T) {
	corefile :=
		`.:0 {
    kubernetes cluster.local 0.0.10.in-addr.arpa {
                endpoint http://localhost:8080
		#endpoint https://kubernetes/ 
		#tls admin.pem admin-key.pem ca.pem
		#tls k8s_auth/client2.crt k8s_auth/client2.key k8s_auth/ca2.crt
		namespaces test-1
		pods disabled
    }
`
	doIntegrationTests(t, corefile, dnsTestCases)
}

func TestKubernetesIntegrationPodsInsecure(t *testing.T) {
	corefile :=
		`.:0 {
    kubernetes cluster.local 0.0.10.in-addr.arpa {
                endpoint http://localhost:8080
		namespaces test-1
		pods insecure
    }
`
	doIntegrationTests(t, corefile, dnsTestCasesPodsInsecure)
}
