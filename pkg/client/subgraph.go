package client

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
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

type WorkflowCheckTaskResult struct {
	TaskName string
	Messages []string
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

func (c *ApolloClient) SubmitSubgraphCheck(ctx context.Context, graphId string, variantName string, subgraphName string, schema string) (string, error) {
	var mutation struct {
		Graph struct {
			Variant struct {
				SubmitSubgraphCheckAsync struct {
					CheckRequestSuccess struct {
						TargetURL  string  `graphql:"targetURL"`
						WorkflowID *string `graphql:"workflowID"`
					} `graphql:"... on CheckRequestSuccess"`
					InvalidInputError struct {
						message string
					} `graphql:"... on InvalidInputError"`
					PermissionError struct {
						message string
					} `graphql:"... on PermissionError"`
					PlanError struct {
						message string
					} `graphql:"... on PlanError"`
				} `graphql:"submitSubgraphCheckAsync(input: $input)"`
			} `graphql:"variant(name: $variantName)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":     graphql.ID(graphId),
		"variantName": graphql.String(variantName),
		"input": SubgraphCheckAsyncInput{
			GraphRef:       graphId + "@" + variantName,
			IsSandbox:      false,
			SubgraphName:   subgraphName,
			ProposedSchema: schema,
			GitContext:     GitContextInput{},
			Config:         HistoricQueryParametersInput{},
		},
	}
	err := c.gqlClient.Mutate(ctx, &mutation, vars)
	if err != nil {
		return "", err
	}
	workFlowId := mutation.Graph.Variant.SubmitSubgraphCheckAsync.CheckRequestSuccess.WorkflowID
	return *workFlowId, nil
}

func (c *ApolloClient) CheckWorkflow(ctx context.Context, graphId string, workflowId string) ([]WorkflowCheckTaskResult, error) {
	var taskResults []WorkflowCheckTaskResult
	type Query struct {
		Graph struct {
			Id            string
			CheckWorkflow struct {
				Status string
				Tasks  []struct {
					Typename             string `graphql:"__typename"`
					Status               string
					OperationsCheckTask  OperationsCheckTask  `graphql:"... on OperationsCheckTask"`
					CompositionCheckTask CompositionCheckTask `graphql:"... on CompositionCheckTask"`
					LintCheckTask        LintCheckTask        `graphql:"... on LintCheckTask"`
				} `json:"tasks"`
			} `graphql:"checkWorkflow(id: $workflowId)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":    graphql.ID(graphId),
		"workflowId": graphql.ID(workflowId),
	}

	for {
		select {
		case <-ctx.Done():
			return taskResults, ctx.Err()
		default:
		}

		var query Query

		err := c.gqlClient.Query(ctx, &query, vars)
		if err != nil {
			return taskResults, err
		}

		status := query.Graph.CheckWorkflow.Status

		switch status {
		case "BLOCKED":
			tflog.Warn(ctx, fmt.Sprintf("Workflow %s blocked from completing", workflowId))
			fallthrough
		case "FAILED":
			tflog.Warn(ctx, fmt.Sprintf("Workflow %s failed to complete", workflowId))
			fallthrough
		case "PASSED":
			tflog.Warn(ctx, fmt.Sprintf("Workflow %s completed", workflowId))
			for _, task := range query.Graph.CheckWorkflow.Tasks {

				// Handle only the tasks with FAILED status
				if task.Status != "FAILED" {
					continue
				}

				taskResult := WorkflowCheckTaskResult{
					TaskName: task.Typename,
					Messages: []string{},
				}

				// TODO: Extract meaningful error messages from other type of task
				switch task.Typename {
				case "CompositionCheckTask":
					for _, error := range task.CompositionCheckTask.Result.Errors {
						taskResult.Messages = append(taskResult.Messages, error.Message)
					}
				default:
					tflog.Warn(ctx, fmt.Sprintf("Extracting error messages from task type: %s is not yet supported", task.Typename))
				}

				taskResults = append(taskResults, taskResult)
			}
			return taskResults, nil
		case "PENDING":
			tflog.Warn(ctx, fmt.Sprintf("Waiting for workflow %s to complete...", workflowId))
		default:
		}

		time.Sleep(2 * time.Second)
	}
}
