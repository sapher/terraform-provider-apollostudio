package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Graph struct {
	Id               string
	Name             string
	Description      string
	GraphType        string
	ReportingEnabled bool
	AccountId        string
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

func (c *ApolloClient) CreateGraph(ctx context.Context, id string, name string, description string) (Graph, error) {
	var mutation struct {
		NewService Graph `graphql:"newService(accountId: $accountId, id: $id, name: $name, description: $description)"`
	}
	vars := map[string]interface{}{
		"accountId":   graphql.ID(c.orgId),
		"id":          graphql.ID(id),
		"name":        graphql.String(name),
		"description": graphql.String(description),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return Graph{}, err
	}
	return mutation.NewService, nil
}

func (c *ApolloClient) RemoveGraph(ctx context.Context, graphId string) error {
	var mutation struct {
		Graph struct {
			Delete string
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId": graphql.ID(graphId),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return err
	}
	return nil
}

// UpdateGraphName updates the name of a graph
// title is a synonym for name.
func (c *ApolloClient) UpdateGraphName(ctx context.Context, graphId string, newName string) error {
	var mutation struct {
		Service struct {
			UpdateTitle struct {
				Id string
			} `graphql:"updateTitle(title: $newTitle)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":  graphql.ID(graphId),
		"newTitle": graphql.String(newName),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return err
	}
	return nil
}

func (c *ApolloClient) UpdateGraphDescription(ctx context.Context, graphId string, newDescription string) error {
	var mutation struct {
		Graph struct {
			UpdateDescription struct {
				Id string
			} `graphql:"updateDescription(description: $newDescription)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":        graphql.ID(graphId),
		"newDescription": graphql.String(newDescription),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return err
	}
	return nil
}
