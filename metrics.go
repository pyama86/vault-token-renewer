package main

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
	"time"
)

var tokenTTLCollector = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "vault_token_renewer",
	Subsystem: "token",
	Name:      "ttl",
})

type VaultTokenTTLSetter struct {
	vault *api.Client
}

func NewVaultTokenTTLSetter(vault *api.Client) *VaultTokenTTLSetter {
	return &VaultTokenTTLSetter{vault: vault}
}

func (s *VaultTokenTTLSetter) Run() {
	for {
		secret, err := s.vault.Auth().Token().Lookup(s.vault.Token())
		if err != nil {
			log.Error(err)
		}
		ttl, err := secret.Data["ttl"].(json.Number).Float64()
		if err != nil {
			log.Error(err)
		}
		tokenTTLCollector.Set(ttl)
		time.Sleep(10 * time.Second)
	}
}

type MetricsServer struct {
}

func NewMetricsServer() *MetricsServer {
	return &MetricsServer{}
}

func (s *MetricsServer) Run() error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":8080", nil)
}
