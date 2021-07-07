package resolver

import (
	log "github.com/sirupsen/logrus"

	"github.com/miekg/dns"
)

type Resolver interface {
	Resolve(fqdn string, dnsType uint16) (*dns.Msg, error)
}

func Resolve(primary Resolver, secondary Resolver, fqdn string, dnsType uint16) (*dns.Msg, error) {
	var resp *dns.Msg
	var err error

	resp, err = primary.Resolve(fqdn, dnsType)
	if err != nil {
		log.Warnln(err, ", trying failback server")

		resp, err = secondary.Resolve(fqdn, dnsType)
		if err != nil {
			return nil, err
		}
	}

	for _, r := range resp.Answer {
		log.Debugf("resolved: %s", r.String())
	}

	return resp, nil
}
