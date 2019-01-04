package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// Config hold configuration
type Config struct {
	Listen            string    `json:"listen"`
	RequireClientName string    `json:"require_client_name"`
	TLS               TLSConfig `json:"tls"`
	ProxyPass         string    `json:"proxy_pass"`
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
		Listen:    ":9101",
		ProxyPass: "http://127.0.0.1:9100",
		TLS: TLSConfig{
			CA:   "ca.pem",
			Cert: "server.pem",
			Key:  "server-key.pem",
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

	if genConfigFlag {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(cfg)
		os.Exit(0)
	}

	if configFileFlag != "" {
		buf, err := ioutil.ReadFile(configFileFlag)
		if err != nil {
			log.Fatal(err)
		}

		if err = json.Unmarshal(buf, &cfg); err != nil {
			log.Fatal(err)
		}
	}

	rpURL, err := url.Parse(cfg.ProxyPass)
	if err != nil {
		log.Fatal(err)
	}
	rp := httputil.NewSingleHostReverseProxy(rpURL)

	promCApem, err := ioutil.ReadFile(cfg.TLS.CA)
	if err != nil {
		log.Fatal(err)
	}

	promCAPool := x509.NewCertPool()
	if ok := promCAPool.AppendCertsFromPEM(promCApem); !ok {
		log.Fatal("can't add cert to client CA pool")
	}

	srv := http.Server{
		Addr:              cfg.Listen,
		Handler:           rp,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      10 * time.Second,
		TLSConfig: &tls.Config{
			ClientAuth:            tls.RequireAndVerifyClientCert,
			ClientCAs:             promCAPool,
			VerifyPeerCertificate: verifyClientCert(cfg),
		},
	}

	log.Printf("listen on %s", cfg.Listen)

	if err = srv.ListenAndServeTLS(cfg.TLS.Cert, cfg.TLS.Key); err != nil {
		log.Fatal(err)
	}
}

func verifyClientCert(cfg Config) func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		for _, asnData := range rawCerts {
			cert, err := x509.ParseCertificate(asnData)
			if err != nil {
				return err
			}

			if cert.IsCA {
				continue
			}

			if cfg.RequireClientName != "" {
				if err = cert.VerifyHostname(cfg.RequireClientName); err != nil {
					return err
				}
			}
		}
		return nil
	}
}
