package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/0w0mewo/umidns/cache"
	"github.com/0w0mewo/umidns/resolver"
	"github.com/miekg/dns"
)

var cfg *Config

func main() {

	// new upstream resolver
	dohclient, backupclient := resolver.NewDoHClient(&http.Client{
		Transport: &http.Transport{
			Proxy: cfg.GetProxyFunc(),
		},
	}, cfg.UpStreamDoH), resolver.NewTCPClient(cfg.UpStreamTcp, true)

	// new cache instance
	recCache := cache.NewMemCache()

	// register dns server handler
	queryHandler := dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {
		var upstreamResp *dns.Msg
		var err error
		var rcode int // RCODE that should reply to client
		ttl := cfg.CacheTTL

		switch m.Opcode {
		// query
		case dns.OpcodeQuery:
			for _, q := range m.Question {
				// do upstream resolving if cache not exist
				if cached := recCache.Get(q.Name); cached != nil {
					upstreamResp = cached.(*dns.Msg)
					rcode = upstreamResp.Rcode

					upstreamResp.SetReply(m)
					upstreamResp.SetRcode(m, rcode) // buggy dns library, rcode should manually set

				} else {
					upstreamResp, err = resolver.Resolve(dohclient, backupclient, q.Name, q.Qtype)
					rcode = upstreamResp.Rcode
					upstreamResp.SetReply(m)
					if err != nil {
						rcode = dns.RcodeServerFailure
						log.Errorln(err)
					}

					// set ttl of cache
					//TODO: OTHER CASES
					switch {
					// domain exist
					case len(upstreamResp.Answer) > 0:
						if cfg.CacheTTL <= 0 {
							ttl = int64(upstreamResp.Answer[0].Header().Ttl)
						}

					// domain not exist or other errors
					case rcode > 0:
						ttl = cache.DefaultTimeout

					default:
						if cfg.CacheTTL <= 0 {
							ttl = cache.DefaultTimeout
						}

					}

					upstreamResp.SetRcode(m, rcode) // buggy dns library, rcode should manually set

					// add to cache
					recCache.Add(q.Name, upstreamResp, time.Duration(ttl)*time.Second)

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

	// register running server instances

	// start and run dns server
	server := &dns.Server{
		Addr:      ":" + strconv.Itoa(cfg.Port),
		Net:       "udp",
		ReusePort: false,
		Handler:   queryHandler,
	}
	go func() {
		log.Infof("server running on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}

	}()

	// wait for shutdown
	<-shutdown
	server.Shutdown()
	log.Infoln("server shutdown")
}

func init() {
	cfg = NewConfig()

	flag.IntVar(&cfg.Port, "port", 53, "port for listening DNS request")
	flag.StringVar(&cfg.ProxyUrl, "proxy", "", "set proxy for passing DoH traffic")
	flag.StringVar(&cfg.UpStreamDoH, "doh", "https://1.1.1.1/dns-query", "set DoH url")
	flag.StringVar(&cfg.UpStreamTcp, "tcp", "8.8.8.8:853", "set tcp dns address, must be addr:port")
	flag.BoolVar(&cfg.Debug, "dbg", false, "turn on debug log")
	flag.Int64Var(&cfg.CacheTTL, "ttl", 0, "set ttl of cache record, in seconds, 0 to set automatically")
	flag.Parse()

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}
}
