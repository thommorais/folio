package pb_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
)

func newShare(t *testing.T, s scenario, label string) domain.Share {
	t.Helper()
	share, err := pb.NewShareRepository(s.app).Create(t.Context(), domain.Share{
		ID: "share0000000001", ProjectID: domain.ProjectID(s.project.Id), IssueID: domain.IssueID(s.ticket.Id),
		Label: label, Token: strings.Repeat("k", 43), CreatedBy: domain.UserID(s.owner.Id),
	})
	if err != nil {
		t.Fatal(err)
	}
	return share
}

func TestShareRepositoryRoundTrip(t *testing.T) {
	s := setup(t)
	repo := pb.NewShareRepository(s.app)
	created := newShare(t, s, "vendor")

	got, err := repo.GetByToken(t.Context(), created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID || got.IssueID != domain.IssueID(s.ticket.Id) || got.Label != "vendor" || got.CreatedBy != domain.UserID(s.owner.Id) {
		t.Errorf("read back %+v, want what was written", got)
	}
	if got.LastAccessedAt != nil || got.CreatedAt.IsZero() {
		t.Errorf("fresh share has last access %v and created %v", got.LastAccessedAt, got.CreatedAt)
	}

	at := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	if err := repo.Touch(t.Context(), created.ID, at); err != nil {
		t.Fatal(err)
	}
	listed, err := repo.List(t.Context(), domain.ProjectID(s.project.Id))
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].LastAccessedAt == nil || !listed[0].LastAccessedAt.Equal(at) {
		t.Errorf("listed %+v, want one share visited at %v", listed, at)
	}

	if err := repo.Delete(t.Context(), created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetByToken(t.Context(), created.Token); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("revoked token err = %v, want not found", err)
	}
}

func TestDeletingTheIssueDeletesItsShares(t *testing.T) {
	s := setup(t)
	share := newShare(t, s, "vendor")

	if err := s.app.Delete(s.ticket); err != nil {
		t.Fatal(err)
	}
	if _, err := pb.NewShareRepository(s.app).GetByID(t.Context(), share.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("share outlived its issue: err = %v", err)
	}
}

func TestShareIsVisibleToMembersOnly(t *testing.T) {
	s := setup(t)
	share := newShare(t, s, "vendor")
	rec, err := s.app.FindRecordById(pb.ColShares, string(share.ID))
	if err != nil {
		t.Fatal(err)
	}

	if !canView(t, s.app, rec, s.viewer) {
		t.Error("a member cannot see the project's shares")
	}
	if canView(t, s.app, rec, s.stranger) {
		t.Error("a stranger can see the project's shares, and so their tokens")
	}
}

// The web app creates shares straight through the collection API, so the
// create rule is the only check on that path.
func TestShareCreateRuleOverRest(t *testing.T) {
	s := setup(t)
	elsewhere := newRecord(t, s.app, pb.ColProjects, map[string]any{
		"slug": "elsewhere", "name": "Elsewhere", "domain": s.other.Id,
	})
	foreign := newRecord(t, s.app, pb.ColIssues, map[string]any{
		"domain": s.other.Id, "project": elsewhere.Id, "kind": "ticket",
		"slug": "foreign", "title": "Foreign", "status": "open", "priority": "low",
	})

	router, err := apis.NewRouter(s.app)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}

	post := func(user *core.Record, body map[string]any) (int, map[string]any) {
		t.Helper()
		token, err := user.NewAuthToken()
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/collections/"+pb.ColShares+"/records", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		var out map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}

	valid := func(by *core.Record) map[string]any {
		return map[string]any{"project": s.project.Id, "issue": s.ticket.Id, "label": "vendor", "created_by": by.Id}
	}

	code, out := post(s.editor, valid(s.editor))
	if code != http.StatusOK {
		t.Fatalf("editor create = %d %v, want 200", code, out)
	}
	if token, _ := out["token"].(string); len(token) != 43 {
		t.Errorf("server generated token %q, want 43 characters", token)
	}

	for _, tc := range []struct {
		name string
		user *core.Record
		body func() map[string]any
	}{
		{"viewer", s.viewer, func() map[string]any { return valid(s.viewer) }},
		{"stranger", s.stranger, func() map[string]any { return valid(s.stranger) }},
		{"filed under someone else", s.editor, func() map[string]any { return valid(s.owner) }},
		{"client-chosen token", s.editor, func() map[string]any {
			b := valid(s.editor)
			b["token"] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
			return b
		}},
		{"issue from another project", s.editor, func() map[string]any {
			b := valid(s.editor)
			b["issue"] = foreign.Id
			return b
		}},
		{"no target", s.editor, func() map[string]any {
			b := valid(s.editor)
			delete(b, "issue")
			return b
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if code, out := post(tc.user, tc.body()); code == http.StatusOK {
				t.Errorf("create accepted: %v", out)
			}
		})
	}
}
