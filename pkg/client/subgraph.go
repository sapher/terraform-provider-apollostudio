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
	Errors         []struct {
		Message string
	}
}

const (
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
	Status   CheckWorkflowTaskStatus
	Details  []WorkflowCheckTaskResultDetail
}

// CheckWorkflowResult is the final outcome of a check workflow.
type CheckWorkflowResult struct {
	Status CheckWorkflowStatus
	Tasks  []WorkflowCheckTaskResult
}

// Failed reports whether the workflow as a whole did not pass.
func (r CheckWorkflowResult) Failed() bool {
	return r.Status != CheckWorkflowStatusPassed
}

// PublishSubGraph publishes the subgraph and returns the composition errors reported by Apollo, if any.
// When composition errors are returned, the subgraph was stored but the supergraph was not updated.
func (c *ApolloClient) PublishSubGraph(ctx context.Context, graphId string, variantName string, name string, schema string, url string, revision string) ([]string, error) {
	var mutation struct {
		Graph *struct {
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
	if err := c.gqlClient.Mutate(ctx, &mutation, vars); err != nil {
		return nil, err
	}
	if mutation.Graph == nil {
		return nil, fmt.Errorf("graph %q: %w", graphId, ErrNotFound)
	}
	compositionErrors := make([]string, 0, len(mutation.Graph.PublishSubGraph.Errors))
	for _, e := range mutation.Graph.PublishSubGraph.Errors {
		compositionErrors = append(compositionErrors, e.Message)
	}
	return compositionErrors, nil
}

func (c *ApolloClient) GetSubGraphs(ctx context.Context, graphId string, variantName string, includeDeleted bool) ([]SubGraph, error) {
	var query struct {
		Graph struct {
			Variant struct {
				SubGraphs []SubGraph `graphql:"subgraphs(includeDeleted: $includeDeleted)"`
			} `graphql:"variant(name: $variantName)"`
		} `graphql:"graph(id: $graphId)"`
	}
	vars := map[string]interface{}{
		"graphId":        graphql.ID(graphId),
		"variantName":    graphql.String(variantName),
		"includeDeleted": graphql.Boolean(includeDeleted),
	}
	err := c.gqlClient.Query(ctx, &query, vars)
	if err != nil {
		return nil, err
	}
	return query.Graph.Variant.SubGraphs, nil
}

func (c *ApolloClient) GetSubGraph(ctx context.Context, graphId string, variantName string, subgraphName string) (SubGraph, error) {
	var query struct {
		Graph *struct {
			Variant *struct {
				SubGraph *SubGraph `graphql:"subgraph(name: $subgraphName)"`
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
	if query.Graph == nil || query.Graph.Variant == nil || query.Graph.Variant.SubGraph == nil {
		return SubGraph{}, ErrNotFound
	}
	return *query.Graph.Variant.SubGraph, nil
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
	return c.gqlClient.Mutate(ctx, &mutation, vars)
}

type checkErrorMessage struct {
	Message string
}

// SubmitSubgraphCheck starts an asynchronous subgraph check and returns its workflow ID.
func (c *ApolloClient) SubmitSubgraphCheck(ctx context.Context, graphId string, variantName string, subgraphName string, schema string) (string, error) {
	var mutation struct {
		Graph *struct {
			Variant *struct {
				SubmitSubgraphCheckAsync struct {
					Typename            string `graphql:"__typename"`
					CheckRequestSuccess struct {
						TargetURL  string  `graphql:"targetURL"`
						WorkflowID *string `graphql:"workflowID"`
					} `graphql:"... on CheckRequestSuccess"`
					InvalidInputError checkErrorMessage `graphql:"... on InvalidInputError"`
					PermissionError   checkErrorMessage `graphql:"... on PermissionError"`
					PlanError         checkErrorMessage `graphql:"... on PlanError"`
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
	if mutation.Graph == nil || mutation.Graph.Variant == nil {
		return "", fmt.Errorf("variant %s@%s: %w", graphId, variantName, ErrNotFound)
	}

	res := mutation.Graph.Variant.SubmitSubgraphCheckAsync
	switch res.Typename {
	case "InvalidInputError":
		return "", fmt.Errorf("invalid check input: %s", res.InvalidInputError.Message)
	case "PermissionError":
		return "", fmt.Errorf("permission denied: %s", res.PermissionError.Message)
	case "PlanError":
		return "", fmt.Errorf("plan error: %s", res.PlanError.Message)
	}
	if res.CheckRequestSuccess.WorkflowID == nil || *res.CheckRequestSuccess.WorkflowID == "" {
		return "", fmt.Errorf("check request returned no workflow ID (response type %q)", res.Typename)
	}
	return *res.CheckRequestSuccess.WorkflowID, nil
}

// CheckWorkflow polls a check workflow until it completes or ctx is done, then returns its outcome.
func (c *ApolloClient) CheckWorkflow(ctx context.Context, graphId string, workflowId string) (CheckWorkflowResult, error) {
	type Query struct {
		Graph struct {
			Id            string
			CheckWorkflow *struct {
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

	for round := 0; ; round++ {
		var query Query

		if err := c.gqlClient.Query(ctx, &query, vars); err != nil {
			return CheckWorkflowResult{}, err
		}
		if query.Graph.CheckWorkflow == nil {
			return CheckWorkflowResult{}, fmt.Errorf("check workflow %q: %w", workflowId, ErrNotFound)
		}

		workflowStatus := query.Graph.CheckWorkflow.Status
		tflog.Debug(ctx, fmt.Sprintf("Workflow %s round %d status: %s", workflowId, round, workflowStatus))

		switch workflowStatus {
		case CheckWorkflowStatusFailed, CheckWorkflowStatusPassed:
			result := CheckWorkflowResult{
				Status: workflowStatus,
				Tasks:  make([]WorkflowCheckTaskResult, 0, len(query.Graph.CheckWorkflow.Tasks)),
			}

			for _, task := range query.Graph.CheckWorkflow.Tasks {
				taskResult := WorkflowCheckTaskResult{
					TaskName: task.Typename,
					Status:   task.Status,
					Details:  make([]WorkflowCheckTaskResultDetail, 0),
				}
				switch task.Typename {
				case TaskTypeCompositionCheck:
					for _, compositionErr := range task.CompositionCheckTask.Result.Errors {
						taskResult.Details = append(taskResult.Details, WorkflowCheckTaskResultDetail{
							Message: compositionErr.Message,
							Level:   LogLevelError,
						})
					}

				case TaskTypeOperationsCheck:
					for _, change := range task.OperationsCheckTask.Result.Changes {
						switch change.Severity {
						case SeverityFailure, SeverityNotice:
							logLevel := LogLevelInfo
							if change.Severity == SeverityFailure {
								logLevel = LogLevelError
							}
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
							srcLocations := make([]string, 0, len(diagnostic.SourceLocations))
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
				}

				// Any failed or blocked task must surface as an error, even when no detail could be extracted.
				if (task.Status == CheckWorkflowTaskStatusFailed || task.Status == CheckWorkflowTaskStatusBlocked) && !hasErrorDetail(taskResult.Details) {
					taskResult.Details = append(taskResult.Details, WorkflowCheckTaskResultDetail{
						Message: fmt.Sprintf("Task %s for unknown reason, please check on apollo studio dashboard for more details", strings.ToLower(string(task.Status))),
						Level:   LogLevelError,
					})
				}

				result.Tasks = append(result.Tasks, taskResult)
			}

			return result, nil
		case CheckWorkflowStatusPending:
			tflog.Info(ctx, fmt.Sprintf("Waiting for workflow %s to complete...", workflowId))
		default:
			tflog.Warn(ctx, fmt.Sprintf("Workflow %s has unexpected status %q, still waiting", workflowId, workflowStatus))
		}

		timer := time.NewTimer(c.pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return CheckWorkflowResult{}, fmt.Errorf("waiting for check workflow %s: %w", workflowId, ctx.Err())
		case <-timer.C:
		}
	}
}

func hasErrorDetail(details []WorkflowCheckTaskResultDetail) bool {
	for _, d := range details {
		if d.Level == LogLevelError {
			return true
		}
	}
	return false
}
