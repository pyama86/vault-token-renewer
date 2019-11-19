package vault

import (
	"context"
	"github.com/hashicorp/vault/api"
	"github.com/lestrrat-go/backoff"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"time"
)

func NewRenewer(vaultAddr, token string, increment string, gracePeriod string) (*Renewer, error) {
	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	client, err := api.NewClient(&api.Config{Address: vaultAddr, HttpClient: httpClient})
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	incrementDuration, err := time.ParseDuration(increment)
	if err != nil {
		return nil, err
	}

	gracePeriodDuration, err := time.ParseDuration(gracePeriod)
	if err != nil {
		return nil, err
	}

	return &Renewer{
		token:       token,
		client:      client,
		increment:   incrementDuration,
		gracePeriod: gracePeriodDuration,
	}, nil
}

type Renewer struct {
	token  string
	client *api.Client

	// increment is time to extend token's TTL.
	increment time.Duration

	// gracePeriod is time to renew before expire.
	gracePeriod time.Duration
}

var policy = backoff.NewExponential(
	backoff.WithInterval(500*time.Millisecond),
	backoff.WithJitterFactor(0.05),
	backoff.WithMaxRetries(25),
)

func (c *Renewer) renew() (*api.Secret, error) {
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

func (c *Renewer) sleepDuration(ttl time.Duration) time.Duration {
	return time.Duration(ttl.Seconds()-c.gracePeriod.Seconds()) * time.Second
}

func (c *Renewer) Run() error {
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
