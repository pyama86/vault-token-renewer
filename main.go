package main

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

var Logger *log.Logger

func init() {
	prometheus.MustRegister(tokenTTLCollector)
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	vault, err := NewVaultClientFromEnv()
	if err != nil {
		return err
	}

	tokenRenewer, err := NewTokenRenewer(
		vault,
		os.Getenv("VAULT_INCREMENT"),
		os.Getenv("VAULT_GRACE_PERIOD"),
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create TokenRenewer")
	}

	go getTokenTTL(vault)

	http.Handle("/metrics", promhttp.Handler())

	go http.ListenAndServe(":8080", nil)

	return tokenRenewer.Run()
}
