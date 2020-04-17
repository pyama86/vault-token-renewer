package main

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"time"
)

var Logger *log.Logger
var tokenTTL = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "vault_token_renewer",
	Subsystem: "token",
	Name:      "ttl",
})

func init() {
	prometheus.MustRegister(tokenTTL)
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	tokenRenewer, err := NewRenewer(
		os.Getenv("VAULT_ADDR"),
		os.Getenv("VAULT_TOKEN"),
		os.Getenv("VAULT_INCREMENT"),
		os.Getenv("VAULT_GRACE_PERIOD"),
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create Renewer")
	}

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	client, err := api.NewClient(&api.Config{Address: os.Getenv("VAULT_ADDR"), HttpClient: httpClient})
	if err != nil {
		return err
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))
	go getTokenTTL(client)

	http.Handle("/metrics", promhttp.Handler())

	go http.ListenAndServe(":8080", nil)

	return tokenRenewer.Run()
}

func getTokenTTL(vault *api.Client) error {
	for {
		secret, err := vault.Auth().Token().Lookup(os.Getenv("VAULT_TOKEN"))
		if err != nil {
			return err
		}
		ttl, err := secret.Data["ttl"].(json.Number).Float64()
		if err != nil {
			return err
		}
		tokenTTL.Set(ttl)
		time.Sleep(10 * time.Second)
	}
	return nil
}
