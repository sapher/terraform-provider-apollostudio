package client

import (
	"context"
	"fmt"
	"strings"
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

var (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

type LogLevel string

type WorkflowCheckTaskResultDetail struct {
	Message string
	Level   LogLevel
}

type WorkflowCheckTaskResult struct {
	TaskName TaskTypename
	Details  []WorkflowCheckTaskResultDetail
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
				Status CheckWorkflowStatus `graphql:"status"`
				Tasks  []struct {
					Typename             TaskTypename            `graphql:"__typename"`
					Status               CheckWorkflowTaskStatus `graphql:"status"`
					OperationsCheckTask  OperationsCheckTask     `graphql:"... on OperationsCheckTask"`
					CompositionCheckTask CompositionCheckTask    `graphql:"... on CompositionCheckTask"`
					LintCheckTask        LintCheckTask           `graphql:"... on LintCheckTask"`
					DownstreamCheckTask  DownstreamCheckTask     `graphql:"... on DownstreamCheckTask"`
					FilterCheckTask      FilterCheckTask         `graphql:"... on FilterCheckTask"`
					ProposalsCheckTask   ProposalsCheckTask      `graphql:"... on ProposalsCheckTask"`
				} `json:"tasks"`
			} `graphql:"checkWorkflow(id: $workflowId)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":    graphql.ID(graphId),
		"workflowId": graphql.ID(workflowId),
	}

	var round = 0

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

		workflowStatus := query.Graph.CheckWorkflow.Status

		tflog.Info(ctx, fmt.Sprintf("Workflow %s round %d status: %s", workflowId, round, workflowStatus))

		switch workflowStatus {
		case CheckWorkflowStatusFailed:
			tflog.Info(ctx, fmt.Sprintf("Workflow %s failed to complete", workflowId))
			fallthrough
		case CheckWorkflowStatusPassed:
			tflog.Info(ctx, fmt.Sprintf("Workflow %s completed", workflowId))

			taskResults = make([]WorkflowCheckTaskResult, 0)

			for _, task := range query.Graph.CheckWorkflow.Tasks {
				taskResult := WorkflowCheckTaskResult{
					TaskName: task.Typename,
					Details:  make([]WorkflowCheckTaskResultDetail, 0),
				}
				switch task.Typename {
				case TaskTypeCompositionCheck:
					for _, error := range task.CompositionCheckTask.Result.Errors {
						taskResult.Details = append(taskResult.Details, WorkflowCheckTaskResultDetail{
							Message: error.Message,
							Level:   LogLevelError,
						})
					}

				case TaskTypeOperationsCheck:
					for _, change := range task.OperationsCheckTask.Result.Changes {
						logLevel := LogLevelInfo
						if change.Severity == SeverityFailure {
							logLevel = LogLevelError
						}

						switch change.Severity {
						case SeverityFailure, SeverityNotice:
							taskResult.Details = append(taskResult.Details, WorkflowCheckTaskResultDetail{
								Message: fmt.Sprintf("%s (severity: %s, code: %s, category: %s)", change.Description, change.Severity, change.Code, change.Category),
								Level:   logLevel,
							})
						default:
							tflog.Warn(ctx, fmt.Sprintf("Change severity: %s is not yet supported", change.Severity))
						}
					}

				case TaskTypeLintCheck:
					for _, diagnostic := range task.LintCheckTask.Result.Diagnostics {
						switch diagnostic.Level {
						case DiagnosticLevelError, DiagnosticLevelWarning:
							logLevel := LogLevelInfo
							if diagnostic.Level == DiagnosticLevelError {
								logLevel = LogLevelError
							}
							var srcLocations []string = make([]string, 0)
							for _, sourceLocation := range diagnostic.SourceLocations {
								srcLocations = append(srcLocations, fmt.Sprintf("line %d-%d col %d-%d", sourceLocation.Start.Line, sourceLocation.End.Line, sourceLocation.Start.Column, sourceLocation.End.Column))
							}
							taskResult.Details = append(taskResult.Details, WorkflowCheckTaskResultDetail{
								Message: fmt.Sprintf("%s - %s (level: %s, rule: %s) %s", diagnostic.Coordinate, diagnostic.Message, diagnostic.Level, diagnostic.Rule, strings.Join(srcLocations, ", ")),
								Level:   logLevel,
							})
						default:
							tflog.Warn(ctx, fmt.Sprintf("Diagnostic level: %s is not yet supported", diagnostic.Level))
						}
					}

				case TaskTypeProposalsCheck, TaskTypeDownstreamCheck, TaskTypeFilterCheck:
					if task.ProposalsCheckTask.Status == CheckWorkflowTaskStatusFailed {
						taskResult.Details = append(taskResult.Details, WorkflowCheckTaskResultDetail{
							Message: "Task failed for unknown reason, please check on apollo studio dashboard for more details",
							Level:   LogLevelError,
						})
					}

				default:
				}

				taskResults = append(taskResults, taskResult)
			}

			return taskResults, nil
		case CheckWorkflowStatusPending:
			tflog.Info(ctx, fmt.Sprintf("Waiting for workflow %s to complete...", workflowId))
		default:
		}

		round++
		time.Sleep(2 * time.Second)
	}
}
