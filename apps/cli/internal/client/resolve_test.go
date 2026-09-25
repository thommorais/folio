package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type call struct {
	method string
	path   string
	body   map[string]any
}

func resolveServer(t *testing.T, patchStatus int) (*Client, *[]call) {
	t.Helper()

	var calls []call
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := call{method: r.Method, path: r.URL.Path}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &c.body)
		}
		calls = append(calls, c)

		switch {
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"e1","kind":"resolution","issue_id":"tk1","body":"why"}`))
		case r.Method == http.MethodPatch && patchStatus != http.StatusOK:
			w.WriteHeader(patchStatus)
			_, _ = w.Write([]byte(`{"message":"refused"}`))
		case r.Method == http.MethodPatch:
			_, _ = w.Write([]byte(`{"id":"tk1","status":"done","resolution":"A graph.","resolution_entry_id":"e1"}`))
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(server.Close)

	return New(server.URL, "tok"), &calls
}

func TestResolveTicketWritesTheDetailThenCloses(t *testing.T) {
	folio, calls := resolveServer(t, http.StatusOK)

	ticket, err := folio.ResolveTicket("pr1", "tk1", Resolution{Answer: "A graph.", Detail: "The tree hides blockers."})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Resolution != "A graph." || ticket.ResolutionEntry != "e1" {
		t.Fatalf("ticket = %+v", ticket)
	}

	if len(*calls) != 2 {
		t.Fatalf("calls = %+v", *calls)
	}
	post, patch := (*calls)[0], (*calls)[1]
	if post.path != "/api/folio/projects/pr1/entries" || post.body["kind"] != "resolution" ||
		post.body["issue_id"] != "tk1" || post.body["body"] != "The tree hides blockers." {
		t.Errorf("post = %+v", post)
	}
	if patch.path != "/api/folio/issues/tk1" || patch.body["status"] != "done" ||
		patch.body["resolution"] != "A graph." || patch.body["resolution_entry_id"] != "e1" {
		t.Errorf("patch = %+v", patch)
	}
}

func TestResolveTicketWithoutDetailOnlyPatches(t *testing.T) {
	folio, calls := resolveServer(t, http.StatusOK)

	if _, err := folio.ResolveTicket("pr1", "tk1", Resolution{Answer: "Out of scope.", Status: "cancelled"}); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 {
		t.Fatalf("calls = %+v", *calls)
	}
	patch := (*calls)[0]
	if patch.body["status"] != "cancelled" || patch.body["resolution"] != "Out of scope." {
		t.Errorf("patch = %+v", patch)
	}
	if _, sent := patch.body["resolution_entry_id"]; sent {
		t.Error("resolution_entry_id sent with no detail written")
	}
}

func TestResolveTicketDeletesTheDetailWhenTheCloseIsRefused(t *testing.T) {
	folio, calls := resolveServer(t, http.StatusBadRequest)

	if _, err := folio.ResolveTicket("pr1", "tk1", Resolution{Answer: "A graph.", Detail: "why"}); err == nil {
		t.Fatal("want the refusal back")
	}
	if len(*calls) != 3 {
		t.Fatalf("calls = %+v", *calls)
	}
	if del := (*calls)[2]; del.method != http.MethodDelete || del.path != "/api/folio/entries/e1" {
		t.Errorf("cleanup = %+v", del)
	}
}

func TestOpenCycleWithMapDeletesTheMapWhenTheLinkIsRefused(t *testing.T) {
	var calls []call
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, call{method: r.Method, path: r.URL.Path})
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/folio/issues/tk1/cycles":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"cy1","phase":"plan"}`))
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"m1","wayfinder":"map"}`))
		case r.Method == http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"refused"}`))
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	if _, _, err := New(server.URL, "tok").OpenCycleWithMap("pr1", "tk1", "Plan it"); err == nil {
		t.Fatal("want the refusal back")
	}
	if last := calls[len(calls)-1]; last.method != http.MethodDelete || last.path != "/api/folio/issues/m1" {
		t.Errorf("cleanup = %+v", last)
	}
}

func TestMeReadsTheSignedInUser(t *testing.T) {
	var gotPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth = r.URL.Path, r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"token":"fresh","record":{"id":"u1","email":"a@b.c"}}`))
	}))
	defer server.Close()

	id, err := New(server.URL, "tok").Me()
	if err != nil {
		t.Fatal(err)
	}
	if id != "u1" {
		t.Errorf("id = %q, want u1", id)
	}
	if gotPath != "/api/collections/users/auth-refresh" || gotAuth == "" {
		t.Errorf("path = %q, auth = %q", gotPath, gotAuth)
	}
}
