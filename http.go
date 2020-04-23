package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var tokenTTLCollector = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "vault_token_renewer",
	Subsystem: "token",
	Name:      "ttl",
})

type HttpServer struct {
	vault *api.Client
}

func NewHttpServer(vault *api.Client) *HttpServer {
	return &HttpServer{vault: vault}
}

func (s *HttpServer) Run() error {
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health/token", &tokenHealthHandler{vault: s.vault})
	return http.ListenAndServe(":8080", nil)
}

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
			if rerr, ok := err.(*api.ResponseError); ok {
				b, err := json.Marshal(rerr)
				if err != nil {
					logrus.Error(err)
				}
				logrus.Error(string(b))
			}
		} else {
			ttl := secret.Data["ttl"]
			fttl, err := ttl.(json.Number).Float64()
			if err != nil {
				logrus.Error(err)
			}
			tokenTTLCollector.Set(fttl)
		}
		time.Sleep(10 * time.Second)
	}
}

type tokenHealth struct {
	StatusCode int
}
type tokenHealthHandler struct {
	vault *api.Client
}

func (h *tokenHealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vreq := h.vault.NewRequest("POST", "/v1/auth/token/lookup")
	if err := vreq.SetJSONBody(map[string]interface{}{
		"token": h.vault.Token(),
	}); err != nil {
		logrus.Error(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	resp, err := h.vault.RawRequestWithContext(ctx, vreq)
	if err != nil {
		logrus.Error(err)
	}
	defer resp.Body.Close()

	b, err := json.Marshal(tokenHealth{StatusCode: resp.StatusCode})
	if err != nil {
		logrus.Error(err)
	}
	fmt.Fprint(w, string(b))
}
