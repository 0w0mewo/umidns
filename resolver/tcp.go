package resolver

import (
	"time"

	"github.com/miekg/dns"
)

type TCPClient struct {
	client   *dns.Client
	upstream string
}

func NewTCPClient(upstream string, useTLS bool) *TCPClient {
	net := "tcp"
	if useTLS {
		net = "tcp-tls"
	}

	return &TCPClient{
		client: &dns.Client{
			Net:     net,
			Timeout: 2 * time.Second,
		},
		upstream: upstream,
	}
}

func (t *TCPClient) Resolve(fqdn string, dnsType uint16) (*dns.Msg, error) {
	request := new(dns.Msg)
	request.SetQuestion(fqdn, dnsType)

	ret, _, err := t.client.Exchange(request, t.upstream)
	return ret, err

}
