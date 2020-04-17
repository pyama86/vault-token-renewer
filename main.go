package main

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	prometheus.MustRegister(tokenTTLCollector)
}

func main() {
	if err := run(); err != nil {
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

	go NewVaultTokenTTLSetter(vault).Run()
	go NewMetricsServer().Run()

	return tokenRenewer.Run()
}
