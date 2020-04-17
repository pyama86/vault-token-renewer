package main

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"time"
)

var tokenTTLCollector = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "vault_token_renewer",
	Subsystem: "token",
	Name:      "ttl",
})

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
		tokenTTLCollector.Set(ttl)
		time.Sleep(10 * time.Second)
	}
	return nil
}
