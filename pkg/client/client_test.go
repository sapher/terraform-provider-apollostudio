package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type recordedRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
	APIKey    string                 `json:"-"`
}

// fakeServer replies to successive GraphQL requests with the given bodies (the last one repeats).
type fakeServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []recordedRequest
}

func newFakeServer(t *testing.T, responses ...string) *fakeServer {
	t.Helper()
	fs := &fakeServer{}
	fs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req recordedRequest
		_ = json.Unmarshal(body, &req)
		req.APIKey = r.Header.Get("x-api-key")

		fs.mu.Lock()
		fs.requests = append(fs.requests, req)
		i := len(fs.requests) - 1
		fs.mu.Unlock()

		if i >= len(responses) {
			i = len(responses) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responses[i]))
	}))
	t.Cleanup(fs.Close)
	return fs
}

func (fs *fakeServer) client() *ApolloClient {
	c := NewClient(fs.URL, "test-key", "test-org")
	c.pollInterval = time.Millisecond
	return c
}

func (fs *fakeServer) requestCount() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return len(fs.requests)
}

func (fs *fakeServer) firstRequest() recordedRequest {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.requests[0]
}

func TestNewClientSendsApiKey(t *testing.T) {
	fs := newFakeServer(t, `{"data":{"me":{"id":"u1","name":"User"}}}`)
	me, err := fs.client().GetMe(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if me.Id != "u1" || me.Name != "User" {
		t.Errorf("unexpected identity: %+v", me)
	}
	if got := fs.firstRequest().APIKey; got != "test-key" {
		t.Errorf("x-api-key header = %q, want test-key", got)
	}
}

func TestGetGraph(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":{"id":"g1","name":"Graph","description":"d","graphType":"CLASSIC","reportingEnabled":true}}}`)
		g, err := fs.client().GetGraph(context.Background(), "g1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g.Id != "g1" || g.Name != "Graph" || !g.ReportingEnabled {
			t.Errorf("unexpected graph: %+v", g)
		}
	})
	t.Run("not found", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":null}}`)
		_, err := fs.client().GetGraph(context.Background(), "missing")
		if !IsNotFound(err) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
	t.Run("api error", func(t *testing.T) {
		fs := newFakeServer(t, `{"errors":[{"message":"boom"}]}`)
		_, err := fs.client().GetGraph(context.Background(), "g1")
		if err == nil || IsNotFound(err) || !strings.Contains(err.Error(), "boom") {
			t.Fatalf("expected API error, got %v", err)
		}
	})
}

func TestGetGraphApiKeyNotFound(t *testing.T) {
	fs := newFakeServer(t, `{"data":{"graph":{"apiKeys":[{"id":"k1","keyName":"a","role":"OBSERVER","token":"x","createdAt":"t"}]}}}`)
	c := fs.client()

	k, err := c.GetGraphApiKey(context.Background(), "g1", "k1")
	if err != nil || k.Id != "k1" {
		t.Fatalf("expected key k1, got %+v, %v", k, err)
	}
	if _, err := c.GetGraphApiKey(context.Background(), "g1", "other"); !IsNotFound(err) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetGraphVariantNotFound(t *testing.T) {
	fs := newFakeServer(t, `{"data":{"variant":null}}`)
	if _, err := fs.client().GetGraphVariant(context.Background(), "g@v"); !IsNotFound(err) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetSubGraph(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":{"variant":{"subgraph":{"name":"s","revision":"3","url":"http://x","activePartialSchema":{"sdl":"type Query{a:Int}","createdAt":"t","isLive":true}}}}}}`)
		sg, err := fs.client().GetSubGraph(context.Background(), "g", "v", "s")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sg.Revision != "3" || sg.ActivePartialSchema.Sdl != "type Query{a:Int}" {
			t.Errorf("unexpected subgraph: %+v", sg)
		}
	})
	for name, body := range map[string]string{
		"graph null":    `{"data":{"graph":null}}`,
		"variant null":  `{"data":{"graph":{"variant":null}}}`,
		"subgraph null": `{"data":{"graph":{"variant":{"subgraph":null}}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			fs := newFakeServer(t, body)
			if _, err := fs.client().GetSubGraph(context.Background(), "g", "v", "s"); !IsNotFound(err) {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
		})
	}
}

func TestGetSubGraphsSendsIncludeDeleted(t *testing.T) {
	fs := newFakeServer(t, `{"data":{"graph":{"variant":{"subgraphs":[{"name":"a"},{"name":"b"}]}}}}`)
	sgs, err := fs.client().GetSubGraphs(context.Background(), "g", "v", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sgs) != 2 {
		t.Errorf("expected 2 subgraphs, got %d", len(sgs))
	}
	req := fs.firstRequest()
	if !strings.Contains(req.Query, "subgraphs(includeDeleted: $includeDeleted)") {
		t.Errorf("query does not pass includeDeleted: %s", req.Query)
	}
	if req.Variables["includeDeleted"] != true {
		t.Errorf("includeDeleted variable = %v, want true", req.Variables["includeDeleted"])
	}
}

func TestPublishSubGraph(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":{"publishSubgraph":{"wasCreated":true,"wasUpdated":false,"updatedGateway":true,"createdAt":"t","errors":[]}}}}`)
		errs, err := fs.client().PublishSubGraph(context.Background(), "g", "v", "s", "type Query{a:Int}", "http://x", "1")
		if err != nil || len(errs) != 0 {
			t.Fatalf("expected success, got %v, %v", errs, err)
		}
	})
	t.Run("composition errors", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":{"publishSubgraph":{"wasCreated":false,"wasUpdated":true,"updatedGateway":false,"createdAt":"t","errors":[{"message":"field conflict"}]}}}}`)
		errs, err := fs.client().PublishSubGraph(context.Background(), "g", "v", "s", "x", "http://x", "1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(errs) != 1 || errs[0] != "field conflict" {
			t.Errorf("unexpected composition errors: %v", errs)
		}
	})
	t.Run("graph not found", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":null}}`)
		if _, err := fs.client().PublishSubGraph(context.Background(), "g", "v", "s", "x", "http://x", "1"); !IsNotFound(err) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestSubmitSubgraphCheck(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":{"variant":{"submitSubgraphCheckAsync":{"__typename":"CheckRequestSuccess","targetURL":"http://studio","workflowID":"wf-1"}}}}}`)
		id, err := fs.client().SubmitSubgraphCheck(context.Background(), "g", "v", "s", "x")
		if err != nil || id != "wf-1" {
			t.Fatalf("expected wf-1, got %q, %v", id, err)
		}
		input, ok := fs.firstRequest().Variables["input"].(map[string]interface{})
		if !ok {
			t.Fatalf("input variable has unexpected type %T", fs.firstRequest().Variables["input"])
		}
		if ref := input["graphRef"]; ref != "g@v" {
			t.Errorf("graphRef = %v, want g@v", ref)
		}
	})
	for typename, want := range map[string]string{
		"PermissionError":   "permission denied: nope",
		"InvalidInputError": "invalid check input: nope",
		"PlanError":         "plan error: nope",
	} {
		t.Run(typename, func(t *testing.T) {
			fs := newFakeServer(t, `{"data":{"graph":{"variant":{"submitSubgraphCheckAsync":{"__typename":"`+typename+`","message":"nope"}}}}}`)
			_, err := fs.client().SubmitSubgraphCheck(context.Background(), "g", "v", "s", "x")
			if err == nil || err.Error() != want {
				t.Fatalf("error = %v, want %q", err, want)
			}
		})
	}
	t.Run("missing workflow id", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":{"variant":{"submitSubgraphCheckAsync":{"__typename":"CheckRequestSuccess","targetURL":"","workflowID":null}}}}}`)
		if _, err := fs.client().SubmitSubgraphCheck(context.Background(), "g", "v", "s", "x"); err == nil {
			t.Fatal("expected an error when workflowID is null")
		}
	})
	t.Run("variant not found", func(t *testing.T) {
		fs := newFakeServer(t, `{"data":{"graph":{"variant":null}}}`)
		if _, err := fs.client().SubmitSubgraphCheck(context.Background(), "g", "v", "s", "x"); !IsNotFound(err) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func workflow(status string, tasks string) string {
	return `{"data":{"graph":{"id":"g","checkWorkflow":{"status":"` + status + `","tasks":[` + tasks + `]}}}}`
}

func TestCheckWorkflowPollsUntilDone(t *testing.T) {
	fs := newFakeServer(t,
		workflow("PENDING", ""),
		workflow("PENDING", ""),
		workflow("PASSED", `{"__typename":"CompositionCheckTask","status":"PASSED","result":{"errors":[]}}`),
	)
	res, err := fs.client().CheckWorkflow(context.Background(), "g", "wf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Failed() || len(res.Tasks) != 1 {
		t.Errorf("unexpected result: %+v", res)
	}
	if n := fs.requestCount(); n != 3 {
		t.Errorf("expected 3 polls, got %d", n)
	}
}

func TestCheckWorkflowExtractsDetails(t *testing.T) {
	tasks := strings.Join([]string{
		`{"__typename":"CompositionCheckTask","status":"FAILED","result":{"errors":[{"message":"cannot compose"}]}}`,
		`{"__typename":"OperationsCheckTask","status":"FAILED","result":{"changes":[
			{"code":"FIELD_REMOVED","description":"removed a","severity":"FAILURE","category":"REMOVAL"},
			{"code":"FIELD_ADDED","description":"added b","severity":"NOTICE","category":"ADDITION"}]}}`,
		`{"__typename":"LintCheckTask","status":"PASSED","result":{"diagnostics":[
			{"coordinate":"Query.a","message":"bad name","level":"WARNING","rule":"FIELD_NAMES","sourceLocations":[{"start":{"line":1,"column":2},"end":{"line":1,"column":5}}]},
			{"coordinate":"Query.b","message":"ignored","level":"IGNORED","rule":"X","sourceLocations":[]}]}}`,
	}, ",")
	fs := newFakeServer(t, workflow("FAILED", tasks))
	res, err := fs.client().CheckWorkflow(context.Background(), "g", "wf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Failed() {
		t.Error("expected workflow to be reported as failed")
	}
	if len(res.Tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(res.Tasks))
	}

	composition := res.Tasks[0].Details
	if len(composition) != 1 || composition[0].Message != "cannot compose" || composition[0].Level != LogLevelError {
		t.Errorf("unexpected composition details: %+v", composition)
	}

	ops := res.Tasks[1].Details
	if len(ops) != 2 || ops[0].Level != LogLevelError || ops[1].Level != LogLevelInfo {
		t.Errorf("unexpected operations details: %+v", ops)
	}

	lint := res.Tasks[2].Details
	if len(lint) != 1 || lint[0].Level != LogLevelInfo || !strings.Contains(lint[0].Message, "line 1-1 col 2-5") {
		t.Errorf("unexpected lint details: %+v", lint)
	}
}

func TestCheckWorkflowFailedTaskWithoutDetails(t *testing.T) {
	for _, typename := range []string{"DownstreamCheckTask", "FilterCheckTask", "ProposalsCheckTask", "CustomCheckTask"} {
		t.Run(typename, func(t *testing.T) {
			fs := newFakeServer(t, workflow("FAILED", `{"__typename":"`+typename+`","status":"FAILED"}`))
			res, err := fs.client().CheckWorkflow(context.Background(), "g", "wf")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d := res.Tasks[0].Details; len(d) != 1 || d[0].Level != LogLevelError {
				t.Errorf("expected one error detail for failed %s, got %+v", typename, d)
			}
		})
	}
	t.Run("blocked", func(t *testing.T) {
		fs := newFakeServer(t, workflow("FAILED", `{"__typename":"DownstreamCheckTask","status":"BLOCKED"}`))
		res, _ := fs.client().CheckWorkflow(context.Background(), "g", "wf")
		if d := res.Tasks[0].Details; len(d) != 1 || d[0].Level != LogLevelError {
			t.Errorf("expected blocked task to produce an error detail, got %+v", d)
		}
	})
}

func TestCheckWorkflowRespectsContext(t *testing.T) {
	fs := newFakeServer(t, workflow("PENDING", ""))
	c := fs.client()
	c.pollInterval = time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.CheckWorkflow(ctx, "g", "wf")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("CheckWorkflow returned after %v, should stop as soon as the context is done", elapsed)
	}
}

func TestCheckWorkflowNotFound(t *testing.T) {
	fs := newFakeServer(t, `{"data":{"graph":{"id":"g","checkWorkflow":null}}}`)
	if _, err := fs.client().CheckWorkflow(context.Background(), "g", "wf"); !IsNotFound(err) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
