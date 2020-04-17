package main

import (
	"github.com/pkg/errors"
	"log"
	"os"
)

var Logger *log.Logger

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

	return tokenRenewer.Run()
}
