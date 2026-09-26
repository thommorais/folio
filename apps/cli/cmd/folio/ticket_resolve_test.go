package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"folio/cli/internal/client"
)

func resolveServer(t *testing.T) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{method: r.Method, path: r.URL.Path}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		got = append(got, req)

		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"e1","kind":"resolution","issue_id":"tk1","body":"why"}`))
		default:
			_, _ = w.Write([]byte(`{"id":"tk1","project_id":"pr1","kind":"ticket","title":"Tree or graph","status":"done","wayfinder":"grilling","resolution":"A graph.","tags":[],"depends_on":[]}`))
		}
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "")

	return &got
}

func runTicket(t *testing.T, stdin string, args ...string) error {
	t.Helper()

	stdout, in := os.Stdout, os.Stdin
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(w, strings.NewReader(stdin))
	_ = w.Close()
	os.Stdout, os.Stdin = devnull, r
	t.Cleanup(func() {
		os.Stdout, os.Stdin = stdout, in
		_ = devnull.Close()
		_ = r.Close()
	})

	cmd := ticketCommand()
	cmd.SetOut(devnull)
	cmd.SetErr(devnull)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func silenced(t *testing.T, fn func() error) error {
	t.Helper()
	stdout := os.Stdout
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = devnull
	defer func() {
		os.Stdout = stdout
		_ = devnull.Close()
	}()
	return fn()
}

func TestTicketResolveClosesWithTheDetail(t *testing.T) {
	got := resolveServer(t)

	if err := runTicket(t, "The tree hides blockers.\n", "resolve", "tk1", "A graph.", "--detail", "-"); err != nil {
		t.Fatal(err)
	}

	if len(*got) != 3 {
		t.Fatalf("requests = %+v", *got)
	}
	lookup, post, patch := (*got)[0], (*got)[1], (*got)[2]
	if lookup.method != http.MethodGet || lookup.path != "/api/folio/issues/tk1" {
		t.Errorf("lookup = %+v", lookup)
	}
	if post.path != "/api/folio/projects/pr1/entries" || post.body["kind"] != "resolution" || post.body["body"] != "The tree hides blockers.\n" {
		t.Errorf("post = %+v", post)
	}
	if patch.method != http.MethodPatch || patch.body["status"] != "done" ||
		patch.body["resolution"] != "A graph." || patch.body["resolution_entry_id"] != "e1" {
		t.Errorf("patch = %+v", patch)
	}
}

func TestTicketResolveCancelRulesItOut(t *testing.T) {
	got := resolveServer(t)

	if err := runTicket(t, "", "resolve", "tk1", "Out of scope.", "--cancel"); err != nil {
		t.Fatal(err)
	}

	if len(*got) != 2 {
		t.Fatalf("requests = %+v", *got)
	}
	if patch := (*got)[1]; patch.body["status"] != "cancelled" || patch.body["resolution"] != "Out of scope." {
		t.Errorf("patch = %+v", patch)
	}
}

func TestBriefShowsEachChildsAnswer(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = stdout })

	err = renderBrief(client.TicketBrief{
		Ticket: client.Ticket{ID: "map1", Title: "The map", Wayfinder: "map"},
		Children: []client.Ticket{
			{ID: "tk1", Kind: "ticket", Status: "done", Priority: "high", Title: "Tree or graph", Resolution: "A graph."},
			{ID: "tk2", Kind: "ticket", Status: "open", Priority: "medium", Title: "Draw the edges"},
		},
	})
	_ = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	out, _ := io.ReadAll(r)

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "tk1") && !strings.HasSuffix(line, "Tree or graph: A graph.") {
			t.Errorf("answered child = %q", line)
		}
		if strings.Contains(line, "tk2") && !strings.HasSuffix(line, "Draw the edges") {
			t.Errorf("open child = %q", line)
		}
	}
}

func TestTicketBodyReadsStdin(t *testing.T) {
	for _, args := range [][]string{
		{"create", "Tree or graph", "--body", "-"},
		{"update", "tk1", "--body", "-"},
	} {
		t.Run(args[0], func(t *testing.T) {
			got := cycleServer(t, "folio")
			if err := runTicket(t, "## Question\n\nTree or graph?\n", args...); err != nil {
				t.Fatal(err)
			}
			last := (*got)[len(*got)-1]
			if last.body["body"] != "## Question\n\nTree or graph?\n" {
				t.Errorf("body = %q", last.body["body"])
			}
		})
	}
}

func meServer(t *testing.T) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{method: r.Method, path: r.URL.Path + "?" + r.URL.RawQuery}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		got = append(got, req)

		switch {
		case r.URL.Path == "/api/collections/users/auth-refresh":
			_, _ = w.Write([]byte(`{"token":"fresh","record":{"id":"u1"}}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/issues"):
			_, _ = w.Write([]byte(`{"issues":[]}`))
		default:
			_, _ = w.Write([]byte(`{"id":"tk1","project_id":"pr1","kind":"ticket","title":"Tree or graph","status":"open","assignee":"u1","tags":[],"depends_on":[]}`))
		}
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	t.Setenv("FOLIO_PROJECT", "folio")

	return &got
}

func TestAssigneeMeResolvesTheSignedInUser(t *testing.T) {
	t.Run("update claims", func(t *testing.T) {
		got := meServer(t)
		if err := runTicket(t, "", "update", "tk1", "--assignee", "me"); err != nil {
			t.Fatal(err)
		}
		last := (*got)[len(*got)-1]
		if last.method != http.MethodPatch || last.body["assignee"] != "u1" {
			t.Errorf("patch = %+v", last)
		}
	})

	t.Run("list filters", func(t *testing.T) {
		got := meServer(t)
		if err := runTicket(t, "", "list", "--assignee", "me"); err != nil {
			t.Fatal(err)
		}
		last := (*got)[len(*got)-1]
		if !strings.Contains(last.path, "assignee=u1") {
			t.Errorf("list = %+v", last)
		}
	})

	t.Run("an explicit id costs no lookup", func(t *testing.T) {
		got := meServer(t)
		if err := runTicket(t, "", "update", "tk1", "--assignee", "u9"); err != nil {
			t.Fatal(err)
		}
		if len(*got) != 1 || (*got)[0].body["assignee"] != "u9" {
			t.Errorf("requests = %+v", *got)
		}
	})
}

func TestBriefShowsTheCurrentPlan(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = stdout })

	err = renderBrief(client.TicketBrief{
		Ticket: client.Ticket{ID: "work", Title: "Make the work legible"},
		Map: &client.MapBrief{
			Ticket:   client.Ticket{ID: "map1", Title: "Plan the graph view"},
			Open:     3,
			Frontier: []client.Ticket{{ID: "tk2", Wayfinder: "research", Title: "How dense"}},
		},
	})
	_ = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	out, _ := io.ReadAll(r)

	for _, want := range []string{"cycle plan", "map1  Plan the graph view  3 open", "next  tk2  research  How dense"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("brief is missing %q:\n%s", want, out)
		}
	}
}
