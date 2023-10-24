package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type GraphVariant struct {
	Identity
}

func (c *ApolloClient) GetGraphVariants(ctx context.Context, graphId string) ([]GraphVariant, error) {
	var query struct {
		Graph struct {
			Variants []GraphVariant
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId": graphql.ID(graphId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, err
	}
	return query.Graph.Variants, nil
}

func (c *ApolloClient) GetGraphVariant(ctx context.Context, variantRef string) (GraphVariant, error) {
	var query struct {
		Variant struct {
			GraphVariant GraphVariant `graphql:"... on GraphVariant"`
		} `graphql:"variant(ref: $ref)"`
	}
	vars := map[string]interface{}{
		"ref": graphql.ID(variantRef),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return GraphVariant{}, err
	}
	return query.Variant.GraphVariant, nil
}
