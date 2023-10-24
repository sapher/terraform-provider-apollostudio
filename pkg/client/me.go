package client

import (
	"context"
)

// Me is a struct that represents the current authenticated user.
type Me Identity

// GetMe returns the current authenticated user.
func (c *ApolloClient) GetMe(ctx context.Context) (Me, error) {
	var query struct {
		Me Me
	}
	err := c.gqlClient.Query(ctx, &query, nil)
	if err != nil {
		return Me{}, err
	}
	return query.Me, nil
}
