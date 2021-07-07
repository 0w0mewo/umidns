package main

import (
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	ProxyUrl    string
	Port        int
	UpStreamDoH string
	UpStreamTcp string
	CacheTTL    int64
	Debug       bool
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) GetProxyFunc() func(r *http.Request) (*url.URL, error) {
	return func(r *http.Request) (*url.URL, error) {
		// no proxy
		if cfg.ProxyUrl == "" {
			return nil, nil
		}

		// try to parse proxy url
		proxy, err := url.Parse(cfg.ProxyUrl)
		if err != nil {
			log.Warnf("%s, no proxy will be used", err)
			return nil, nil
		}

		return proxy, nil
	}
}
