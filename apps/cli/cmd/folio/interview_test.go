package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
)

type rawRequest struct {
	method string
	path   string
	body   string
}

const interviewJSON = `{"id":"iv1","project_id":"pr1","issue_id":"tk1","topic":"Tree or graph","agent_status":"waiting","handled":2,
"state":{"terms":[],"questions":[
	{"id":"q1","round":1,"title":"Tree or graph","status":"answered","deps":[],"options":[],"rec":{"why":"x"},"thread":[]},
	{"id":"q2","round":2,"title":"What do we call it","status":"open","deps":[],"options":[],"rec":{"why":"x"},"thread":[]},
	{"id":"q3","round":2,"title":"Who answers","status":"reopened","deps":[],"options":[],"rec":{"why":"x"},"thread":[]}
]},"created_at":"2026-09-26T10:00:00Z","updated_at":"2026-09-26T10:00:00Z"}`

func interviewServer(t *testing.T, pending string) (*[]rawRequest, string) {
	t.Helper()

	var got []rawRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		got = append(got, rawRequest{method: r.Method, path: r.URL.Path, body: string(raw)})

		switch {
		case r.URL.Path == "/api/folio/issues/tk1" || r.URL.Path == "/api/folio/projects/geral/issues/tree-or-graph":
			_, _ = w.Write([]byte(`{"id":"tk1","project_id":"pr1","slug":"tree-or-graph","kind":"ticket","title":"Tree or graph","status":"open","wayfinder":"grilling","tags":[],"depends_on":[]}`))
		case r.URL.Path == "/api/folio/projects/pr1":
			_, _ = w.Write([]byte(`{"id":"pr1","domain_id":"do1","slug":"geral","name":"Geral","members":[]}`))
		case r.URL.Path == "/api/folio/domains":
			_, _ = w.Write([]byte(`{"domains":[{"id":"do0","client_slug":"other","slug":"x","members":[]},{"id":"do1","client_id":"cl1","client_slug":"journ","slug":"shed","members":[]}]}`))
		case strings.HasSuffix(r.URL.Path, "/interview/pending"):
			_, _ = w.Write([]byte(pending))
		case strings.HasSuffix(r.URL.Path, "/interview/finish"):
			_, _ = w.Write([]byte(`{"id":"iv1","issue_id":"tk1","topic":"Tree or graph","finished_at":"2026-09-26T11:00:00Z","state":{"terms":[],"questions":[]}}`))
		case strings.HasSuffix(r.URL.Path, "/interviews"):
			_, _ = w.Write([]byte(`{"interviews":[` + interviewJSON + `]}`))
		case r.Method == http.MethodPatch:
			_, _ = w.Write([]byte(`{"round":3,"added":2,"answered":1,"handled":7,"interview":` + interviewJSON + `}`))
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(interviewJSON))
		default:
			_, _ = w.Write([]byte(interviewJSON))
		}
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "")
	flagJSON = false
	t.Cleanup(func() { flagJSON = false })

	return &got, server.URL
}

func runInterview(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()

	out, _, err := runInterviewErr(t, stdin, args...)
	return out, err
}

func runInterviewErr(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()

	stdout, in := os.Stdout, os.Stdin
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(inW, strings.NewReader(stdin))
	_ = inW.Close()
	os.Stdout, os.Stdin = outW, inR

	cmd := interviewCommand()
	var errOut strings.Builder
	cmd.SetOut(io.Discard)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	runErr := cmd.Execute()

	_ = outW.Close()
	os.Stdout, os.Stdin = stdout, in
	printed, _ := io.ReadAll(outR)
	_ = outR.Close()
	_ = inR.Close()
	return string(printed), errOut.String(), runErr
}

func paths(got []rawRequest) []string {
	out := make([]string, 0, len(got))
	for _, r := range got {
		out = append(out, r.method+" "+r.path)
	}
	return out
}

func TestInterviewStartPrintsTheLinkFirst(t *testing.T) {
	got, url := interviewServer(t, "")
	out, err := runInterview(t, "", "start", "tk1")
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if want := url + "/journ/shed/geral/tickets/tree-or-graph/interview"; lines[0] != want {
		t.Fatalf("first line = %q, want %q", lines[0], want)
	}
	if !strings.Contains(out, "Tree or graph") {
		t.Errorf("start prints the summary, got %q", out)
	}
	last := (*got)[len(*got)-1]
	if last.method != http.MethodPost || last.path != "/api/folio/issues/tk1/interview" {
		t.Errorf("requests = %v", paths(*got))
	}
}

func TestInterviewResolvesATicketSlug(t *testing.T) {
	got, _ := interviewServer(t, "")
	t.Setenv("FOLIO_PROJECT", "geral")
	if _, err := runInterview(t, "", "show", "tree-or-graph"); err != nil {
		t.Fatal(err)
	}
	if (*got)[0].path != "/api/folio/projects/geral/issues/tree-or-graph" {
		t.Errorf("requests = %v", paths(*got))
	}
	if (*got)[1].path != "/api/folio/issues/tk1/interview" {
		t.Errorf("requests = %v", paths(*got))
	}
}

func TestInterviewPatchSendsStdinAsIs(t *testing.T) {
	got, _ := interviewServer(t, "")
	patch := `{"questions":[{"id":"q1","status":"answered"}],"agent":{"handled":7}}`
	out, err := runInterview(t, patch+"\n", "patch", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "round 3: 2 questions added, 1 answered, handled 7" {
		t.Errorf("patch printed %q", out)
	}
	if len(*got) != 1 {
		t.Fatalf("requests = %v", paths(*got))
	}
	req := (*got)[0]
	if req.method != http.MethodPatch || req.path != "/api/folio/issues/tk1/interview" {
		t.Errorf("request = %s %s", req.method, req.path)
	}
	var sent, want any
	_ = json.Unmarshal([]byte(req.body), &sent)
	_ = json.Unmarshal([]byte(patch), &want)
	sentJSON, _ := json.Marshal(sent)
	wantJSON, _ := json.Marshal(want)
	if string(sentJSON) != string(wantJSON) {
		t.Errorf("body = %s, want %s", req.body, patch)
	}
}

func TestInterviewPatchRefusesAnEmptyStdin(t *testing.T) {
	got, _ := interviewServer(t, "")
	if _, err := runInterview(t, "  \n", "patch", "tk1"); err == nil {
		t.Fatal("an empty patch must be refused")
	}
	if len(*got) != 0 {
		t.Errorf("nothing should reach the API, got %v", paths(*got))
	}
	if _, err := runInterview(t, "{not json", "patch", "tk1"); err == nil {
		t.Fatal("a patch that is not JSON must be refused")
	}
}

func TestInterviewPendingPrintsOneLinePerSend(t *testing.T) {
	interviewServer(t, `{"handled":2,"sends":[
		{"seq":3,"at":"2026-09-26T10:01:00Z","actions":[{"type":"answer","q":"q2","kind":"text","text":"interview"}]},
		{"seq":4,"at":"2026-09-26T10:02:00Z","actions":[{"type":"defer","q":"q3"}]}
	]}`)
	out, err := runInterview(t, "", "pending", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("pending printed %q", out)
	}
	var first struct {
		Seq     int              `json:"seq"`
		At      string           `json:"at"`
		Actions []map[string]any `json:"actions"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("line 1 is not JSON: %q", lines[0])
	}
	if first.Seq != 3 || first.At == "" || first.Actions[0]["text"] != "interview" {
		t.Errorf("line 1 = %+v", first)
	}
}

func TestInterviewPendingSaysWhenNothingWasSent(t *testing.T) {
	interviewServer(t, `{"handled":7,"sends":[]}`)
	out, err := runInterview(t, "", "pending", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "nothing sent since handled 7" {
		t.Errorf("pending printed %q", out)
	}
}

func TestInterviewShowListsTheOpenQuestions(t *testing.T) {
	_, url := interviewServer(t, "")
	out, err := runInterview(t, "", "show", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Tree or graph", "round 2", "handled 2", "waiting", "q2", "What do we call it", "q3", "Who answers", url + "/journ/shed/geral/tickets/tree-or-graph/interview"} {
		if !strings.Contains(out, want) {
			t.Errorf("show is missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "q1") {
		t.Errorf("an answered question is not open:\n%s", out)
	}
}

func TestInterviewListReadsEveryInterview(t *testing.T) {
	got, _ := interviewServer(t, "")
	out, err := runInterview(t, "", "list", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	if (*got)[0].path != "/api/folio/issues/tk1/interviews" || !strings.Contains(out, "iv1") {
		t.Errorf("requests = %v, printed %q", paths(*got), out)
	}
}

func TestInterviewFinishSendsTheAnswerAndTheDoc(t *testing.T) {
	got, _ := interviewServer(t, "")
	if _, err := runInterview(t, "# Tree or graph\n\nA graph.\n", "finish", "tk1", "A graph.", "--doc", "-"); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Fatalf("requests = %v", paths(*got))
	}
	req := (*got)[0]
	var body map[string]string
	_ = json.Unmarshal([]byte(req.body), &body)
	if req.method != http.MethodPost || req.path != "/api/folio/issues/tk1/interview/finish" || body["answer"] != "A graph." || body["doc"] != "# Tree or graph\n\nA graph.\n" {
		t.Errorf("request = %s %s %v", req.method, req.path, body)
	}
}

func TestInterviewFinishNeedsTheDoc(t *testing.T) {
	got, _ := interviewServer(t, "")
	if _, err := runInterview(t, "", "finish", "tk1", "A graph."); err == nil {
		t.Fatal("finish without --doc must be refused")
	}
	if len(*got) != 0 {
		t.Errorf("nothing should reach the API, got %v", paths(*got))
	}
}

func TestInterviewAliases(t *testing.T) {
	for _, want := range []string{"grill", "wayfinder"} {
		if !slices.Contains(interviewCommand().Aliases, want) {
			t.Errorf("folio interview is aliased folio %s", want)
		}
	}
}

func TestInterviewGuidesOnStderrAndKeepsStdoutClean(t *testing.T) {
	interviewServer(t, "")
	out, guide, err := runInterviewErr(t, "", "start", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "next:") {
		t.Errorf("stdout must stay parseable, got %q", out)
	}
	if !strings.Contains(guide, "next:") || !strings.Contains(guide, "folio interview pending tk1") {
		t.Errorf("stderr guide = %q", guide)
	}
}

func TestInterviewJSONPrintsNoGuide(t *testing.T) {
	interviewServer(t, "")
	flagJSON = true
	_, guide, err := runInterviewErr(t, "", "show", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	if guide != "" {
		t.Errorf("--json prints no guide, got %q", guide)
	}
}

func TestInterviewPendingGuideNamesTheLastSeq(t *testing.T) {
	interviewServer(t, `{"handled":2,"sends":[{"seq":3,"at":"a","actions":[]},{"seq":5,"at":"b","actions":[]}]}`)
	_, guide, err := runInterviewErr(t, "", "pending", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(guide, `"handled":5`) {
		t.Errorf("guide = %q", guide)
	}
}

func TestInterviewPendingGuideStopsWhenNothingWasSent(t *testing.T) {
	interviewServer(t, `{"handled":2,"sends":[]}`)
	_, guide, err := runInterviewErr(t, "", "pending", "tk1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(guide, "stop") {
		t.Errorf("guide = %q", guide)
	}
}
