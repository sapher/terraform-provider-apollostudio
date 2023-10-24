package client

import (
	"net/http"

	"github.com/hasura/go-graphql-client"
)

type ApolloClient struct {
	orgId     string
	gqlClient *graphql.Client
}

type Identity struct {
	Id   string
	Name string
}

func NewClient(host string, apiKey string, orgId string) *ApolloClient {
	return &ApolloClient{
		orgId: orgId,
		gqlClient: graphql.NewClient(host, http.DefaultClient).WithRequestModifier(func(r *http.Request) {
			r.Header.Set("x-api-key", apiKey)
		}),
	}
}
