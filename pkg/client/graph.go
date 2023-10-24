package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Graph struct {
	Title            string
	Description      string
	GraphType        string
	ReportingEnabled bool
	AccountId        string
	Identity
}

func (c *ApolloClient) GetGraphs(ctx context.Context) ([]Graph, error) {
	var query struct {
		Organization struct {
			Graphs []Graph
		} `graphql:"organization(id: $orgId)"`
	}
	vars := map[string]interface{}{
		"orgId": graphql.ID(c.orgId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, err
	}
	return query.Organization.Graphs, nil
}

func (c *ApolloClient) GetGraph(ctx context.Context, graphId string) (Graph, error) {
	var query struct {
		Graph Graph `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId": graphql.ID(graphId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return Graph{}, err
	}
	return query.Graph, nil
}
