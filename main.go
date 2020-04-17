package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"log"
	"os"
)

var Logger *log.Logger

func main() {
	app := cli.NewApp()
	app.Name = "vault-token-renewer"
	app.Action = func(c *cli.Context) error {
		tokenRenewer, err := NewRenewer(
			os.Getenv("VAULT_ADDR"),
			os.Getenv("VAULT_TOKEN"),
			os.Getenv("VAULT_INCREMENT"),
			os.Getenv("VAULT_GRACE_PERIOD"),
		)
		if err != nil {
			return errors.Wrap(err, "Failed to create Renewer")
		}

		tokenRenewer.Run()

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
