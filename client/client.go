package client

import (
	"context"
	"net/http"

	"github.com/hasura/go-graphql-client"
)

type ApolloClient struct {
	orgId     string
	gqlClient *graphql.Client
}

type Me struct {
	Id   string
	Name string
}

type Organization struct {
	Id               string
	Name             string
	IsOnTrial        bool
	IsOnExpiredTrial bool
	IsLocked         bool
}

type Graph struct {
	Id               string
	Name             string
	Title            string
	Description      string
	GraphType        string
	ReportingEnabled bool
	AccountId        string
}

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

type GraphVariant struct {
	Id                  string
	Name                string
	HasSupergraphSchema bool
}

type Identity struct {
	Id   string
	Name string
}

type GraphApiKey struct {
	Id        string
	KeyName   string
	Role      string
	Token     string
	CreatedAt string
	CreatedBy Identity
}

func NewClient(host string, apiKey string, orgId string) *ApolloClient {
	return &ApolloClient{
		orgId: orgId,
		gqlClient: graphql.NewClient(host, http.DefaultClient).WithRequestModifier(func(r *http.Request) {
			r.Header.Set("x-api-key", apiKey)
		}),
	}
}

func (c *ApolloClient) GetMe(ctx context.Context) (Me, error) {
	var query struct {
		Me struct {
			Id   string
			Name string
			User struct {
				FullName string
			} `graphql:"... on User"`
		}
	}
	err := c.gqlClient.Query(ctx, &query, nil)
	if err != nil {
		return Me{}, err
	}
	return Me{
		Id:   query.Me.Id,
		Name: query.Me.Name,
	}, nil
}

func (c *ApolloClient) GetOrganization(ctx context.Context) (Organization, error) {
	var query struct {
		Organization struct {
			Id   string
			Name string
		} `graphql:"organization(id: $orgId)"`
	}
	vars := map[string]interface{}{
		"orgId": graphql.ID(c.orgId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return Organization{}, err
	}
	return Organization{
		Id:   query.Organization.Id,
		Name: query.Organization.Name,
	}, nil
}

func (c *ApolloClient) GetGraphs(ctx context.Context) ([]Graph, error) {
	var query struct {
		Organization struct {
			Graphs []struct {
				Id   string
				Name string
			}
		} `graphql:"organization(id: $orgId)"`
	}
	vars := map[string]interface{}{
		"orgId": graphql.ID(c.orgId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, err
	}
	graphs := make([]Graph, 0)
	for _, g := range query.Organization.Graphs {
		graphs = append(graphs, Graph{
			Id:   g.Id,
			Name: g.Name,
		})
	}
	return graphs, nil
}

func (c *ApolloClient) GetGraph(ctx context.Context, graphId string) (Graph, error) {
	var query struct {
		Graph struct {
			Id               string
			Name             string
			Title            string
			Description      string
			GraphType        string
			ReportingEnabled bool
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId": graphql.ID(graphId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return Graph{}, err
	}
	return Graph{
		Id:               query.Graph.Id,
		Name:             query.Graph.Name,
		Title:            query.Graph.Title,
		Description:      query.Graph.Description,
		GraphType:        query.Graph.GraphType,
		ReportingEnabled: query.Graph.ReportingEnabled,
	}, nil
}

func (c *ApolloClient) GetGraphVariant(ctx context.Context, variantRef string) (GraphVariant, error) {
	var query struct {
		Variant struct {
			GraphVariant struct {
				Id                  string
				Name                string
				HasSupergraphSchema bool
			} `graphql:"... on GraphVariant"`
		} `graphql:"variant(ref: $ref)"`
	}
	vars := map[string]interface{}{
		"ref": graphql.ID(variantRef),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return GraphVariant{}, err
	}
	return GraphVariant{
		Id:                  query.Variant.GraphVariant.Id,
		Name:                query.Variant.GraphVariant.Name,
		HasSupergraphSchema: query.Variant.GraphVariant.HasSupergraphSchema,
	}, nil
}

func (c *ApolloClient) GetGraphVariants(ctx context.Context, graphId string) ([]GraphVariant, error) {
	var query struct {
		Graph struct {
			Variants []struct {
				Id                  string
				Name                string
				HasSupergraphSchema bool
			}
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId": graphql.ID(graphId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, err
	}
	graphVariants := make([]GraphVariant, 0)
	for _, v := range query.Graph.Variants {
		graphVariants = append(graphVariants, GraphVariant{
			Id:                  v.Id,
			Name:                v.Name,
			HasSupergraphSchema: v.HasSupergraphSchema,
		})
	}
	return graphVariants, nil
}

func (c *ApolloClient) GetGraphApiKeys(ctx context.Context, graphId string) ([]GraphApiKey, error) {
	var query struct {
		Graph struct {
			ApiKeys []struct {
				Id        string
				KeyName   string
				Role      string
				CreatedAt string
				Token     string
				CreatedBy struct {
					Id   string
					Name string
				}
			}
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId": graphql.ID(graphId),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, err
	}
	graphApiKeys := make([]GraphApiKey, 0)
	for _, ak := range query.Graph.ApiKeys {
		graphApiKeys = append(graphApiKeys, GraphApiKey{
			Id:        ak.Id,
			KeyName:   ak.KeyName,
			Role:      ak.Role,
			CreatedAt: ak.CreatedAt,
			Token:     ak.Token,
			CreatedBy: Identity{
				Id:   ak.CreatedBy.Id,
				Name: ak.CreatedBy.Name,
			},
		})
	}
	return graphApiKeys, nil
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
			NewKey struct {
				Id        string
				KeyName   string
				Role      string
				CreatedAt string
				Token     string
				CreatedBy struct {
					Id   string
					Name string
				}
			} `graphql:"newKey(keyName: $keyName)"`
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
	return GraphApiKey{
		Id:        mutation.Graph.NewKey.Id,
		KeyName:   mutation.Graph.NewKey.KeyName,
		Role:      mutation.Graph.NewKey.Role,
		CreatedAt: mutation.Graph.NewKey.CreatedAt,
		Token:     mutation.Graph.NewKey.Token,
		CreatedBy: Identity{
			Id:   mutation.Graph.NewKey.CreatedBy.Id,
			Name: mutation.Graph.NewKey.CreatedBy.Name,
		},
	}, nil
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

func (c *ApolloClient) GetSubGraphs(ctx context.Context, graphId string, variantName string, includeDeleted bool) ([]SubGraph, error) {
	var query struct {
		Graph struct {
			Variant struct {
				SubGraphs []struct {
					Name                string
					Url                 string
					Revision            string
					ActivePartialSchema struct {
						Sdl       string
						CreatedAt string
						IsLive    bool
					}
				} `graphql:"subgraphs"`
			} `graphql:"variant(name: $variantName)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":     graphql.ID(graphId),
		"variantName": graphql.String(variantName),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	subGraphs := make([]SubGraph, 0)
	if err != nil {
		return subGraphs, err
	}
	for _, sg := range query.Graph.Variant.SubGraphs {
		subGraphs = append(subGraphs, SubGraph{
			Name:     sg.Name,
			Revision: sg.Revision,
			Url:      sg.Url,
			ActivePartialSchema: PartialSchema{
				Sdl:       sg.ActivePartialSchema.Sdl,
				CreatedAt: sg.ActivePartialSchema.CreatedAt,
				IsLive:    sg.ActivePartialSchema.IsLive,
			},
		})
	}
	return subGraphs, nil
}

func (c *ApolloClient) GetSubGraph(ctx context.Context, graphId string, variantName string, subgraphName string) (SubGraph, error) {
	subgraphs, err := c.GetSubGraphs(ctx, graphId, variantName, true)
	if err != nil {
		return SubGraph{}, err
	}
	for _, sg := range subgraphs {
		if sg.Name == subgraphName {
			return sg, nil
		}
	}
	return SubGraph{}, nil
}

func (c *ApolloClient) PublishSubGraph(ctx context.Context, graphId string, variantName string, subgraphName string, revision string, sdl string, url string) error {
	var mutation struct {
		Graph struct {
			PublishSubgraph struct {
				WasCreated bool
				WasUpdated bool
			} `graphql:"publishSubgraph(graphVariant: $variantName, name: $subgraphName, revision: $revision, activePartialSchema: { sdl: $sdl }, url: $url)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":      graphql.ID(graphId),
		"variantName":  graphql.String(variantName),
		"subgraphName": graphql.String(subgraphName),
		"revision":     graphql.String(revision),
		"sdl":          graphql.String(sdl),
		"url":          graphql.String(url),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return err
	}
	return nil
}

func (c *ApolloClient) RemoveSubGraph(ctx context.Context, graphId string, variantName string, subgraphName string) error {
	var mutation struct {
		Graph struct {
			RemoveImplementingServiceAndTriggerComposition struct {
				Errors []struct {
					Message string
					Code    string
				}
			} `graphql:"removeImplementingServiceAndTriggerComposition(graphVariant: $variantName, name: $subgraphName)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":      graphql.ID(graphId),
		"variantName":  graphql.String(variantName),
		"subgraphName": graphql.String(subgraphName),
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return err
	}
	return nil
}
