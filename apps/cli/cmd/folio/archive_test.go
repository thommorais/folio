package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestTodoUpdateArchivesAndRestores(t *testing.T) {
	cases := map[string]bool{"--archive": true, "--unarchive": false}

	for flag, want := range cases {
		t.Run(flag, func(t *testing.T) {
			got := todoServer(t, pendingTodo)

			if err := run(t, "update", "t1", flag); err != nil {
				t.Fatal(err)
			}

			if len(*got) != 1 {
				t.Fatalf("requests = %d, want 1", len(*got))
			}
			req := (*got)[0]
			if req.method != http.MethodPatch || req.path != "/api/folio/issues/t1" {
				t.Errorf("request = %s %s", req.method, req.path)
			}
			if archived, ok := req.body["archived"].(bool); !ok || archived != want {
				t.Errorf("archived = %v, want %v", req.body["archived"], want)
			}
		})
	}
}

func TestTodoUpdateRejectsArchiveWithUnarchive(t *testing.T) {
	got := todoServer(t, pendingTodo)

	err := run(t, "update", "t1", "--archive", "--unarchive")
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error = %v, want a mutually exclusive complaint", err)
	}
	if len(*got) != 0 {
		t.Errorf("sent %d requests, want none", len(*got))
	}
}

func TestTodoListArchivedAsksForTheArchivedView(t *testing.T) {
	got := todoServer(t, `{"issues":[]}`)
	t.Setenv("FOLIO_PROJECT", "folio")

	if err := run(t, "list", "--archived"); err != nil {
		t.Fatal(err)
	}

	if q := (*got)[len(*got)-1].query; !strings.Contains(q, "archived=only") {
		t.Errorf("query = %q, want archived=only", q)
	}
}

func TestTodoListHidesArchivedByDefault(t *testing.T) {
	got := todoServer(t, `{"issues":[]}`)
	t.Setenv("FOLIO_PROJECT", "folio")

	if err := run(t, "list"); err != nil {
		t.Fatal(err)
	}

	if q := (*got)[len(*got)-1].query; strings.Contains(q, "archived") {
		t.Errorf("query = %q, want no archived parameter", q)
	}
}

func TestTicketUpdateArchivesAndRestores(t *testing.T) {
	cases := map[string]bool{"--archive": true, "--unarchive": false}

	for flag, want := range cases {
		t.Run(flag, func(t *testing.T) {
			got := resolveServer(t)

			if err := runTicket(t, "", "update", "tk1", flag); err != nil {
				t.Fatal(err)
			}

			patch := (*got)[len(*got)-1]
			if patch.method != http.MethodPatch {
				t.Fatalf("last request = %s %s, want a patch", patch.method, patch.path)
			}
			if archived, ok := patch.body["archived"].(bool); !ok || archived != want {
				t.Errorf("archived = %v, want %v", patch.body["archived"], want)
			}
		})
	}
}

func TestTicketUpdateRejectsArchiveWithUnarchive(t *testing.T) {
	got := resolveServer(t)

	err := runTicket(t, "", "update", "tk1", "--archive", "--unarchive")
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error = %v, want a mutually exclusive complaint", err)
	}
	if len(*got) != 0 {
		t.Errorf("sent %d requests, want none", len(*got))
	}
}

func TestTicketListArchivedAsksForTheArchivedView(t *testing.T) {
	got := resolveServer(t)
	t.Setenv("FOLIO_PROJECT", "folio")

	_ = runTicket(t, "", "list", "--archived")

	if len(*got) == 0 || !strings.Contains((*got)[len(*got)-1].query, "archived=only") {
		t.Errorf("requests = %+v, want a list with archived=only", *got)
	}
}
