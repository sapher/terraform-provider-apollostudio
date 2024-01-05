package client

type HistoricQueryParametersInput struct{}

type GitContextInput struct {
	Branch    *string `json:"branch"`
	Commit    *string `json:"commit"`
	Committer *string `json:"committer"`
	Message   *string `json:"message"`
	RemoteUrl *string `json:"remoteUrl"`
}

type SubgraphCheckAsyncInput struct {
	Config         HistoricQueryParametersInput `json:"config"`
	GitContext     GitContextInput              `json:"gitContext"`
	GraphRef       string                       `json:"graphRef"`
	IsSandbox      bool                         `json:"isSandbox"`
	ProposedSchema string                       `json:"proposedSchema"`
	SubgraphName   string                       `json:"subgraphName"`
}

type FieldChangeSummaryCounts struct {
	Additions int `json:"additions"`
	Removals  int `json:"removals"`
	Edits     int `json:"edits"`
}

type ChangeSummary struct {
	Field FieldChangeSummaryCounts `json:"field"`
	Total FieldChangeSummaryCounts `json:"total"`
	Type  FieldChangeSummaryCounts `json:"type"`
}

type OperationsCheckResult struct {
	Id                         string        `json:"id"`
	ChangeSummary              ChangeSummary `json:"changeSummary"`
	NumberOfAffectedOperations int           `json:"numberOfAffectedOperations"`
	NumberOfCheckedOperations  int           `json:"numberOfCheckedOperations"`
}

type OperationsCheckTask struct {
	Id     string                  `json:"id"`
	Result OperationsCheckResult   `json:"result"`
	Status CheckWorkflowTaskStatus `json:"status"`
}

type SchemaCompositionError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CompositionResult struct {
	Errors []SchemaCompositionError `json:"errors"`
}

type CompositionCheckTask struct {
	Id                 string                  `json:"id"`
	Result             CompositionResult       `json:"result"`
	Status             CheckWorkflowTaskStatus `json:"status"`
	CoreSchemaModified bool                    `json:"coreSchemaModified"`
}

type LintDiagnostic struct {
	Coordinate string `json:"coordinate"`
	Message    string `json:"message"`
}

type LintResult struct {
	Diagnostics []LintDiagnostic `json:"diagnostics"`
}

type LintCheckTask struct {
	Id     string `json:"id"`
	Result LintResult
}

const (
	StatusFailed  CheckWorkflowTaskStatus = "FAILED"
	StatusPassed  CheckWorkflowTaskStatus = "PASSED"
	StatusBlocked CheckWorkflowTaskStatus = "BLOCKED"
	StatusPending CheckWorkflowTaskStatus = "PENDING"
)

type CheckWorkflowTaskStatus string
