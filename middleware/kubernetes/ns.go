package kubernetes

import (
	"net"
	"strings"

	"github.com/coredns/coredns/middleware/etcd/msg"

	"github.com/miekg/dns"
	"k8s.io/client-go/1.5/pkg/api"
)

const defaultNSName = "ns.dns."

var corednsRecord dns.A

type interfaceAddrser interface {
	interfaceAddrs() ([]net.Addr, error)
}

type interfaceAddrs struct{}

func (i interfaceAddrs) interfaceAddrs() ([]net.Addr, error) {
	return net.InterfaceAddrs()
}

func (k *Kubernetes) recordsForNS(r recordRequest, svcs *[]msg.Service) error {
	ns := k.coreDNSRecord()
	s := msg.Service{
		Host: ns.A.String(),
		Key:  msg.Path(strings.Join([]string{ns.Hdr.Name, r.zone}, "."), "coredns")}
	*svcs = append(*svcs, s)
	return nil
}

// DefaultNSMsg returns an msg.Service representing an A record for
// ns.dns.[zone] -> dns service ip. This A record is needed to legitimize
// the SOA response in middleware.NS(), which is hardcoded at ns.dns.[zone].
func (k *Kubernetes) defaultNSMsg(r recordRequest) msg.Service {
	ns := k.coreDNSRecord()
	s := msg.Service{
		Key:  msg.Path(strings.Join([]string{defaultNSName, r.zone}, "."), "coredns"),
		Host: ns.A.String(),
	}
	return s
}

func isDefaultNS(name string, r recordRequest) bool {
	return strings.Index(name, defaultNSName) == 0 && strings.Index(name, r.zone) == len(defaultNSName)
}

func (k *Kubernetes) coreDNSRecord() dns.A {
	var localIP net.IP
	var svcName string
	var svcNamespace string
	var dnsIP net.IP

	if len(corednsRecord.Hdr.Name) == 0 || corednsRecord.A == nil {
		// get local Pod IP
		addrs, _ := k.interfaceAddrs.interfaceAddrs()

		for _, addr := range addrs {
			ip, _, _ := net.ParseCIDR(addr.String())
			ip = ip.To4()

			if ip == nil || ip.IsLoopback() {
				continue
			}
			localIP = ip
			break
		}
		// Find endpoint matching IP to get service and namespace
		endpointsList := k.APIConn.EndpointsList()

	FindEndpoint:
		for _, ep := range endpointsList.Items {
			for _, eps := range ep.Subsets {
				for _, addr := range eps.Addresses {
					if localIP.Equal(net.ParseIP(addr.IP)) {

						svcNamespace = ep.ObjectMeta.Namespace
						svcName = ep.ObjectMeta.Name
						break FindEndpoint
					}
				}
			}
		}

		if len(svcName) == 0 {
			corednsRecord.Hdr.Name = defaultNSName
			corednsRecord.A = localIP
			return corednsRecord
		}
		// Find service to get ClusterIP
		serviceList := k.APIConn.ServiceList()
	FindService:
		for _, svc := range serviceList {
			if svcName == svc.Name && svcNamespace == svc.Namespace {
				if svc.Spec.ClusterIP == api.ClusterIPNone {
					dnsIP = localIP
				} else {
					dnsIP = net.ParseIP(svc.Spec.ClusterIP)
				}
				break FindService
			}
		}
		if dnsIP == nil {
			dnsIP = localIP
		}

		corednsRecord.Hdr.Name = strings.Join([]string{svcName, svcNamespace, "svc."}, ".")
		corednsRecord.A = dnsIP
	}
	return corednsRecord
}
