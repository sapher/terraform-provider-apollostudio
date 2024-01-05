package client

import (
	"context"
)

type Identity struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (c *ApolloClient) GetMe(ctx context.Context) (Identity, error) {
	var query struct {
		Me Identity
	}
	err := c.gqlClient.Query(ctx, &query, nil)
	if err != nil {
		return Identity{}, err
	}
	return query.Me, nil
}
