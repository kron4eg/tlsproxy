package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// Config hold configuration
type Config struct {
	Listen             string         `json:"listen"`
	RequiredClientName string         `json:"required_client_name"`
	TLS                TLSConfig      `json:"tls"`
	VHosts             map[string]int `json:"vhosts"`
}

// TLSConfig hold TLS configuration
type TLSConfig struct {
	CA   string `json:"ca"`
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

// NewConfig initialize Config with default values
func NewConfig() Config {
	return Config{
		Listen: ":9101",
		TLS: TLSConfig{
			CA:   "ca.pem",
			Cert: "server.pem",
			Key:  "server-key.pem",
		},
		VHosts: map[string]int{
			"": 9100,
		},
	}
}

var (
	configFileFlag string
	genConfigFlag  bool
)

func main() {
	flag.StringVar(&configFileFlag, "config", "", "path to config.json")
	flag.BoolVar(&genConfigFlag, "genconfig", false, "write default config to stdout and exit")
	flag.Parse()

	cfg := NewConfig()

	if configFileFlag != "" {
		buf, err := ioutil.ReadFile(configFileFlag)
		if err != nil {
			log.Fatal(err)
		}

		if err = json.Unmarshal(buf, &cfg); err != nil {
			log.Fatal(err)
		}
	}

	if genConfigFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(cfg)
		os.Exit(0)
	}

	vhostsMap := map[string]http.Handler{}

	for vhost, upstreamPort := range cfg.VHosts {
		rpURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", upstreamPort))
		if err != nil {
			log.Fatal(err)
		}

		rp := httputil.NewSingleHostReverseProxy(rpURL)
		rp.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		}
		vhostsMap[vhost] = rp
	}

	caPem, err := ioutil.ReadFile(cfg.TLS.CA)
	if err != nil {
		log.Fatal(err)
	}

	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caPem); !ok {
		log.Fatal("can't add cert to client CA pool")
	}

	srv := http.Server{
		Addr:              cfg.Listen,
		Handler:           VHostRouter(vhostsMap),
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      10 * time.Second,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  caPool,
		},
	}

	if cfg.RequiredClientName != "" {
		srv.TLSConfig.VerifyPeerCertificate = verifyClientCert(cfg)
	}

	log.Printf("listen on %s", cfg.Listen)

	if err = srv.ListenAndServeTLS(cfg.TLS.Cert, cfg.TLS.Key); err != nil {
		log.Fatal(err)
	}
}

func verifyClientCert(cfg Config) func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		asnData := rawCerts[0]
		cert, err := x509.ParseCertificate(asnData)
		if err != nil {
			return err
		}

		return cert.VerifyHostname(cfg.RequiredClientName)
	}
}

// VHostRouter router
func VHostRouter(vhosts map[string]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next, ok := vhosts[r.Host]
		if !ok {
			next, _ = vhosts[""]
		}
		next.ServeHTTP(w, r)
	})
}
