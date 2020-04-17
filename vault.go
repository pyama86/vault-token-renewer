package main

import (
	"context"
	"github.com/hashicorp/vault/api"
	"github.com/lestrrat-go/backoff"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"time"
)

func NewVaultClientFromEnv() (*api.Client, error) {
	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	client, err := api.NewClient(&api.Config{Address: os.Getenv("VAULT_ADDR"), HttpClient: httpClient})
	if err != nil {
		return nil, err
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))
	return client, nil
}

func NewTokenRenewer(client *api.Client, increment string, gracePeriod string) (*TokenRenewer, error) {
	incrementDuration, err := time.ParseDuration(increment)
	if err != nil {
		return nil, err
	}

	gracePeriodDuration, err := time.ParseDuration(gracePeriod)
	if err != nil {
		return nil, err
	}

	return &TokenRenewer{
		client:      client,
		increment:   incrementDuration,
		gracePeriod: gracePeriodDuration,
	}, nil
}

type TokenRenewer struct {
	client *api.Client

	// increment is time to extend token's TTL.
	increment time.Duration

	// gracePeriod is time to renew before expire.
	gracePeriod time.Duration
}

func (c *TokenRenewer) Run() error {
	for {
		renewedSecret, err := c.renew()
		if err != nil {
			log.Printf("Failed to c.renew() (%s)", err)
			time.Sleep(10 * time.Second)
			continue
		}
		newTTL, err := renewedSecret.TokenTTL()
		if err != nil {
			log.Printf("Failed to renewedSecret.TokenTTL() (%s)", err)
			time.Sleep(10 * time.Second)
			continue
		}
		log.Printf("Success renew, ttl: %s", newTTL)

		sleepDuration := c.sleepDuration(newTTL)

		log.Printf("Sleep %v", sleepDuration)
		time.Sleep(sleepDuration)
	}
}

var policy = backoff.NewExponential(
	backoff.WithInterval(500*time.Millisecond),
	backoff.WithJitterFactor(0.05),
	backoff.WithMaxRetries(25),
)

func (c *TokenRenewer) renew() (*api.Secret, error) {
	b, cancel := policy.Start(context.Background())
	defer cancel()

	for backoff.Continue(b) {
		renewal, err := c.client.Auth().Token().RenewSelf(int(c.increment.Seconds()))
		if err == nil {
			return renewal, nil
		}
		log.Printf("Failed to renew, waiting to retry... (%s)\n", err)
	}
	return nil, errors.New("Failed retring to renew token.")
}

func (c *TokenRenewer) sleepDuration(ttl time.Duration) time.Duration {
	return time.Duration(ttl.Seconds()-c.gracePeriod.Seconds()) * time.Second
}
