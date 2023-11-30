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
	Name                string
	Revision            string
	Url                 string
	ActivePartialSchema PartialSchema
}

type PublishSubGraph struct {
	WasCreated     bool
	WasUpdated     bool
	UpdatedGateway bool
	CreatedAt      string
}

func (c *ApolloClient) PublishSubGraph(ctx context.Context, graphId string, variantName string, name string, schema string, url string, revision string) error {
	var mutation struct {
		Graph struct {
			PublishSubGraph PublishSubGraph `graphql:"publishSubgraph(graphVariant: $variantName, name: $name, activePartialSchema: { sdl: $schema }, url: $url, revision: $revision)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":     graphql.ID(graphId),
		"variantName": graphql.String(variantName),
		"schema":      graphql.String(schema),
		"name":        graphql.String(name),
		"url":         graphql.String(url),
		"revision":    graphql.String(revision),
	}
	return c.gqlClient.Mutate(ctx, &mutation, vars)
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

func (c *ApolloClient) GetSubGraph(ctx context.Context, graphId string, variantName string, subgraphName string) (SubGraph, error) {
	var query struct {
		Graph struct {
			Variant struct {
				SubGraph SubGraph `graphql:"subgraph(name: $subgraphName)"`
			} `graphql:"variant(name: $variantName)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":      graphql.ID(graphId),
		"variantName":  graphql.String(variantName),
		"subgraphName": graphql.ID(subgraphName),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return SubGraph{}, err
	}
	return query.Graph.Variant.SubGraph, nil
}

func (c *ApolloClient) RemoveSubGraph(ctx context.Context, graphId string, variantName string, subgraphName string) error {
	var mutation struct {
		Graph struct {
			RemoveImplementingServiceAndTriggerComposition struct {
				DidExist bool
			} `graphql:"removeImplementingServiceAndTriggerComposition(graphVariant: $variantName, name: $name)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":     graphql.ID(graphId),
		"variantName": graphql.String(variantName),
		"name":        graphql.String(subgraphName),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return err
	}
	return nil
}
