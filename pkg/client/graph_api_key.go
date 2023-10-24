package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type GraphApiKey struct {
	Id        string
	KeyName   string
	Role      string
	Token     string
	CreatedAt string
	CreatedBy Identity
}

func (c *ApolloClient) GetGraphApiKeys(ctx context.Context, graphId string) ([]GraphApiKey, error) {
	var query struct {
		Graph struct {
			ApiKeys []GraphApiKey
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId": graphql.ID(graphId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, err
	}
	return query.Graph.ApiKeys, nil
}

func (c *ApolloClient) GetGraphApiKey(ctx context.Context, graphId string, apiKeyId string) (GraphApiKey, error) {
	graphApiKeys, err := c.GetGraphApiKeys(ctx, graphId)
	if err != nil {
		return GraphApiKey{}, err
	}
	for _, ak := range graphApiKeys {
		if ak.Id == apiKeyId {
			return ak, nil
		}
	}
	return GraphApiKey{}, nil
}

func (c *ApolloClient) CreateGraphApiKey(ctx context.Context, graphId string, keyName string) (GraphApiKey, error) {
	var mutation struct {
		Graph struct {
			NewKey GraphApiKey `graphql:"newKey(keyName: $keyName)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId": graphql.ID(graphId),
		"keyName": graphql.String(keyName),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return GraphApiKey{}, err
	}
	return mutation.Graph.NewKey, nil
}

func (c *ApolloClient) RenameGraphApiKey(ctx context.Context, graphId string, apiKeyId string, newKeyName string) error {
	var mutation struct {
		Graph struct {
			RenameKey struct {
				Id      string
				KeyName string
			} `graphql:"renameKey(id: $apiKeyId, newKeyName: $keyName)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":  graphql.ID(graphId),
		"apiKeyId": graphql.ID(apiKeyId),
		"keyName":  graphql.String(newKeyName),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return err
	}
	return nil
}

func (c *ApolloClient) RemoveGraphApiKey(ctx context.Context, graphId string, apiKeyId string) error {
	var mutation struct {
		Graph struct {
			RemoveKey string `graphql:"removeKey(id: $apiKeyId)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":  graphql.ID(graphId),
		"apiKeyId": graphql.ID(apiKeyId),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return err
	}
	return nil
}
