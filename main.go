package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/0w0mewo/umidns/resolver"
	"github.com/miekg/dns"
)

type Config struct {
	ProxyUrl    string
	Port        int
	UpStreamDoH string
	UpStreamTcp string
}

var cfg Config

func main() {

	// new dns server
	server := &dns.Server{
		Addr: ":" + strconv.Itoa(cfg.Port),
		Net:  "udp",
	}

	// new upstream resolver
	dohclient := resolver.NewDoHClient(&http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				// no proxy
				if cfg.ProxyUrl == "" {
					return nil, nil
				}

				// try to parse proxy url
				proxy, err := url.Parse(cfg.ProxyUrl)
				if err != nil {
					log.Printf("%s, no proxy will be used", err)
					return nil, nil
				}

				return proxy, nil
			},
		},
	}, cfg.UpStreamDoH)

	backupclient := resolver.NewTCPClient(cfg.UpStreamTcp, true)

	// register dns server handler
	server.Handler = dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {
		var upstreamResp *dns.Msg = &dns.Msg{}
		var err error

		switch m.Opcode {
		// query
		case dns.OpcodeQuery:
			for _, q := range m.Question {
				upstreamResp, err = resolver.Resolve(dohclient, backupclient, q.Name, q.Qtype)

				// directly copy what DoH client recv to dns client
				upstreamResp.SetReply(m)

				// set NXDOMAIN status to dns client if empty answer
				if len(upstreamResp.Answer) <= 0 {
					upstreamResp.SetRcode(m, dns.RcodeNameError)
				}

				if err != nil {
					upstreamResp.SetRcode(m, dns.RcodeServerFailure)
					log.Println(err)
				}

				rw.WriteMsg(upstreamResp)
			}

		// not implemented
		default:
			upstreamResp = &dns.Msg{}
			upstreamResp.SetReply(m)
			upstreamResp.SetRcode(m, dns.RcodeNotImplemented)

			rw.WriteMsg(upstreamResp)
		}

	})

	// register kill signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// start and run dns server
	go func() {
		log.Println("server running on " + server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	// wait for shutdown
	<-shutdown
	server.Shutdown()
	log.Println("server shutdown")

}

func init() {
	flag.IntVar(&cfg.Port, "port", 53, "port for listening DNS request")
	flag.StringVar(&cfg.ProxyUrl, "proxy", "", "set proxy for passing DoH traffic")
	flag.StringVar(&cfg.UpStreamDoH, "doh", "https://cloudflare-dns.com/dns-query", "set DoH url")
	flag.StringVar(&cfg.UpStreamTcp, "tcp", "8.8.8.8:853", "set tcp dns address, must be addr:port")

	flag.Parse()
}
