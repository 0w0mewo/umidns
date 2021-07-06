package resolver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/miekg/dns"
)

const USERAGENT = "5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.0.0 Safari/537.36"

type DoHClient struct {
	client *http.Client
	doh    string
}

func NewDoHClient(client *http.Client, doh string) *DoHClient {
	if client == nil {
		return &DoHClient{
			client: &http.Client{
				Timeout: 3 * time.Second,
			},
			doh: doh,
		}
	}

	return &DoHClient{
		client: client,
		doh:    doh,
	}
}

func (d *DoHClient) SetDoH(doh string) {
	d.doh = doh
}

func (d *DoHClient) Resolve(fqdn string, dnsType uint16) (*dns.Msg, error) {
	request := new(dns.Msg)
	request.SetQuestion(fqdn, dnsType)

	// pack dns message to binary so that it can be sent as body of http POST
	bodyToSend, err := request.Pack()
	if err != nil {
		return nil, err
	}

	// make a request
	req, err := http.NewRequest("POST", d.doh, bytes.NewReader(bodyToSend))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("User-Agent", USERAGENT)

	// send request
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	// expected 200
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http return code: %d", resp.StatusCode)
	}

	// read body
	bodyRecv, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// unpack
	ret := &dns.Msg{}
	err = ret.Unpack(bodyRecv)

	return ret, err
}
