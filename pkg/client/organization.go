package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Organization struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (c *ApolloClient) GetOrganization(ctx context.Context) (Organization, error) {
	var query struct {
		Organization `graphql:"organization(id: $orgId)"`
	}
	vars := map[string]interface{}{
		"orgId": graphql.ID(c.orgId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return Organization{}, err
	}
	return query.Organization, nil
}
