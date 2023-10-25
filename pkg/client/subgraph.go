package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type PartialSchema struct {
	Sdl       string
	CreatedAt string
	IsLive    bool
}

type SubGraph struct {
	Name         string
	Revision     string
	Url          string
	ActiveSchema PartialSchema
}

func (c *ApolloClient) GetSubGraphs(ctx context.Context, graphId string, variantName string, includeDeleted bool) ([]SubGraph, error) {
	var query struct {
		Graph struct {
			Variant struct {
				SubGraphs []SubGraph `graphql:"subgraphs"`
			} `graphql:"variant(name: $variantName)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":     graphql.ID(graphId),
		"variantName": graphql.String(variantName),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return make([]SubGraph, 0), err
	}
	return query.Graph.Variant.SubGraphs, nil
}
