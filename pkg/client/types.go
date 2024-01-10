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

type AffectedQuery struct {
	AlreadyApproved bool   `json:"alreadyApproved"`
	AlreadyIgnored  bool   `json:"alreadyIgnored"`
	DisplayName     string `json:"displayName"`
	Id              string `json:"id"`
	IsValid         bool   `json:"isValid"`
	Name            string `json:"name"`
	Signature       string `json:"signature"`
}

type Change struct {
	Code        string         `json:"code"`
	Description string         `json:"description"`
	Severity    ChangeSeverity `json:"severity"`
	Category    ChangeCategory `json:"category"`
}

type OperationsCheckResult struct {
	Id                         string          `json:"id"`
	AffectedQueries            []AffectedQuery `json:"affectedQueries"`
	Changes                    []Change        `json:"changes"`
	ChangeSummary              ChangeSummary   `json:"changeSummary"`
	NumberOfAffectedOperations int             `json:"numberOfAffectedOperations"`
	NumberOfCheckedOperations  int             `json:"numberOfCheckedOperations"`
}

type SourceLocation struct {
	Column int `json:"column"`
	Line   int `json:"line"`
}

type SchemaCompositionError struct {
	Code      string           `json:"code"`
	Message   string           `json:"message"`
	Locations []SourceLocation `json:"locations"`
}

type CompositionResult struct {
	Errors []SchemaCompositionError `json:"errors"`
}

type Coordinate struct {
	ByteOffset int `json:"byteOffset"`
	Column     int `json:"column"`
	Line       int `json:"line"`
}

type Location struct {
	End          Coordinate `json:"end"`
	Start        Coordinate `json:"start"`
	SubgraphName string     `json:"subgraphName"`
}

type LintDiagnostic struct {
	Coordinate      string              `json:"coordinate"`
	Message         string              `json:"message"`
	Level           LintDiagnosticLevel `json:"level"`
	Rule            string              `json:"rule"`
	SourceLocations []Location          `json:"sourceLocations"`
}

type LintStats struct {
	ErrorsCount   int `json:"errorsCount"`
	IgnoredCount  int `json:"ignoredCount"`
	TotalCount    int `json:"totalCount"`
	WarningsCount int `json:"warningsCount"`
}

type LintResult struct {
	Diagnostics []LintDiagnostic `json:"diagnostics"`
	Stats       LintStats        `json:"stats"`
}

type DownstreamCheckResult struct {
	Blocking              bool   `json:"blocking"`
	DownstreamVariantName string `json:"downsteamVariantName"`
	FailsUpstreamWorkflow bool   `json:"failsUpstreamWorkflow"`
}

const (
	SeverityFailure ChangeSeverity = "FAILURE"
	SeverityNotice  ChangeSeverity = "NOTICE"
)

type ChangeSeverity string

const (
	CheckWorkflowStatusFailed  CheckWorkflowStatus = "FAILED"
	CheckWorkflowStatusPassed  CheckWorkflowStatus = "PASSED"
	CheckWorkflowStatusPending CheckWorkflowStatus = "PENDING"
)

type CheckWorkflowStatus string

const (
	CheckWorkflowTaskStatusFailed  CheckWorkflowTaskStatus = "FAILED"
	CheckWorkflowTaskStatusPassed  CheckWorkflowTaskStatus = "PASSED"
	CheckWorkflowTaskStatusBlocked CheckWorkflowTaskStatus = "BLOCKED"
	CheckWorkflowTaskStatusPending CheckWorkflowTaskStatus = "PENDING"
)

type CheckWorkflowTaskStatus string

const (
	DiagnosticLevelError   LintDiagnosticLevel = "ERROR"
	DiagnosticLevelWarning LintDiagnosticLevel = "WARNING"
	DiagnosticLevelIgnored LintDiagnosticLevel = "IGNORED"
)

type LintDiagnosticLevel string

const (
	CategoryAddition    ChangeCategory = "ADDITION"
	CatergoryRemoval    ChangeCategory = "REMOVAL"
	CategoryEdit        ChangeCategory = "EDIT"
	CategoryDeprecation ChangeCategory = "DEPRECATION"
)

type ChangeCategory string

const (
	TaskTypeOperationsCheck  TaskTypename = "OperationsCheckTask"
	TaskTypeCompositionCheck TaskTypename = "CompositionCheckTask"
	TaskTypeLintCheck        TaskTypename = "LintCheckTask"
	TaskTypeDownstreamCheck  TaskTypename = "DownstreamCheckTask"
	TaskTypeProposalsCheck   TaskTypename = "ProposalsCheckTask"
	TaskTypeFilterCheck      TaskTypename = "FilterCheckTask"
)

type TaskTypename string

type OperationsCheckTask struct {
	// Id string `json:"id"`
	Result OperationsCheckResult   `json:"result"`
	Status CheckWorkflowTaskStatus `json:"status"`
}

type CompositionCheckTask struct {
	Result CompositionResult       `json:"result"`
	Status CheckWorkflowTaskStatus `json:"status"`
}

type LintCheckTask struct {
	Result LintResult
	Status CheckWorkflowTaskStatus `json:"status"`
}

type DownstreamCheckTask struct {
	Status  CheckWorkflowTaskStatus `json:"status"`
	Results []DownstreamCheckResult `json:"results"`
}

type ProposalsCheckTask struct {
	Status CheckWorkflowTaskStatus `json:"status"`
}

type FilterCheckTask struct {
	Status CheckWorkflowTaskStatus `json:"status"`
}
