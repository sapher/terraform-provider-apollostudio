package client

import (
	"errors"
	"net/http"
	"time"

	"github.com/hasura/go-graphql-client"
)

// ErrNotFound is returned when the requested object does not exist (or is not visible with the current API key).
var ErrNotFound = errors.New("not found")

// IsNotFound reports whether err (or any error it wraps) is ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// defaultHTTPTimeout bounds every request made to the Apollo API so a stuck call cannot hang Terraform.
const defaultHTTPTimeout = 60 * time.Second

// defaultPollInterval is the delay between two polls of a check workflow.
const defaultPollInterval = 2 * time.Second

type ApolloClient struct {
	orgId        string
	gqlClient    *graphql.Client
	pollInterval time.Duration
}

func NewClient(host string, apiKey string, orgId string) *ApolloClient {
	httpClient := &http.Client{Timeout: defaultHTTPTimeout}
	return &ApolloClient{
		orgId:        orgId,
		pollInterval: defaultPollInterval,
		gqlClient: graphql.NewClient(host, httpClient).WithRequestModifier(func(r *http.Request) {
			r.Header.Set("x-api-key", apiKey)
		}),
	}
}
