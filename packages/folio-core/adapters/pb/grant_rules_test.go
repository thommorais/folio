package pb_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/pb"
)

func canUpdate(t *testing.T, app core.App, record, user *core.Record) bool {
	t.Helper()

	ok, err := app.CanAccessRecord(record, info(user, "PATCH"), record.Collection().UpdateRule)
	if err != nil {
		t.Fatalf("update rule on %s: %v", record.Collection().Name, err)
	}
	return ok
}

type grantCast struct {
	boss, clientMember, teamEditor, teamViewer *core.Record
}

func castGrants(t *testing.T, s scenario) grantCast {
	t.Helper()
	c := grantCast{
		boss:         newUser(t, s.app, "boss@test.local"),
		clientMember: newUser(t, s.app, "member@test.local"),
		teamEditor:   newUser(t, s.app, "team-editor@test.local"),
		teamViewer:   newUser(t, s.app, "team-viewer@test.local"),
	}
	newRecord(t, s.app, pb.ColClientMembers, map[string]any{"client": s.client.Id, "user": c.boss.Id, "role": "owner"})
	newRecord(t, s.app, pb.ColClientMembers, map[string]any{"client": s.client.Id, "user": c.clientMember.Id, "role": "member"})

	readers := newRecord(t, s.app, pb.ColDomains, map[string]any{"client": s.client.Id, "slug": "readers", "name": "Readers"})
	newRecord(t, s.app, pb.ColMembers, map[string]any{"domain": s.other.Id, "project": s.project.Id, "user": c.teamEditor.Id, "role": "viewer"})
	newRecord(t, s.app, pb.ColMembers, map[string]any{"domain": readers.Id, "project": s.project.Id, "user": c.teamViewer.Id, "role": "owner"})
	newRecord(t, s.app, pb.ColProjectGrants, map[string]any{"project": s.project.Id, "domain": s.other.Id, "role": "editor"})
	newRecord(t, s.app, pb.ColProjectGrants, map[string]any{"project": s.project.Id, "domain": readers.Id, "role": "viewer"})
	return c
}

func TestContentRulesFollowTheEffectiveRole(t *testing.T) {
	s := setup(t)
	c := castGrants(t, s)

	for _, tc := range []struct {
		name         string
		user         *core.Record
		read, update bool
	}{
		{"client owner", c.boss, true, true},
		{"plain client member", c.clientMember, false, false},
		{"member of a domain granted editor", c.teamEditor, true, true},
		{"member of a domain granted viewer, beside an editor grant", c.teamViewer, true, false},
		{"personal viewer beside personal editors", s.viewer, true, false},
		{"stranger", s.stranger, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, record := range []*core.Record{s.ticket, s.entry} {
				name := record.Collection().Name
				if got := canView(t, s.app, record, tc.user); got != tc.read {
					t.Errorf("read %s = %v, want %v", name, got, tc.read)
				}
				if got := canUpdate(t, s.app, record, tc.user); got != tc.update {
					t.Errorf("update %s = %v, want %v", name, got, tc.update)
				}
			}
		})
	}
}

func TestProjectAdminRuleFollowsTheEffectiveRole(t *testing.T) {
	s := setup(t)
	c := castGrants(t, s)

	for _, tc := range []struct {
		name  string
		user  *core.Record
		admin bool
	}{
		{"client owner", c.boss, true},
		{"personal owner", s.owner, true},
		{"personal editor", s.editor, false},
		{"member of a domain granted editor", c.teamEditor, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := canUpdate(t, s.app, s.project, tc.user); got != tc.admin {
				t.Errorf("update project = %v, want %v", got, tc.admin)
			}
		})
	}
}

func TestContainersAreVisibleThroughAnyGrant(t *testing.T) {
	s := setup(t)
	c := castGrants(t, s)

	for _, tc := range []struct {
		name   string
		user   *core.Record
		client bool
		domain bool
	}{
		{"client owner", c.boss, true, true},
		{"plain client member", c.clientMember, true, false},
		{"personal grant only", s.viewer, true, true},
		{"stranger", s.stranger, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := canView(t, s.app, s.client, tc.user); got != tc.client {
				t.Errorf("view client = %v, want %v", got, tc.client)
			}
			if got := canView(t, s.app, s.domain, tc.user); got != tc.domain {
				t.Errorf("view project's domain = %v, want %v", got, tc.domain)
			}
		})
	}
}

func TestIssueCreateOverRestFollowsTheDomainGrant(t *testing.T) {
	s := setup(t)
	c := castGrants(t, s)

	router, err := apis.NewRouter(s.app)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}

	post := func(user *core.Record, slug string) int {
		t.Helper()
		token, err := user.NewAuthToken()
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(map[string]any{
			"domain": s.domain.Id, "project": s.project.Id, "kind": "ticket",
			"slug": slug, "title": slug, "status": "open", "priority": "low",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/collections/"+pb.ColIssues+"/records", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := post(c.teamEditor, "by-editor"); code != http.StatusOK {
		t.Errorf("member of a domain granted editor: create = %d, want 200", code)
	}
	if code := post(c.teamViewer, "by-viewer"); code == http.StatusOK {
		t.Error("member of a domain granted viewer created an issue")
	}
	if code := post(c.boss, "by-boss"); code != http.StatusOK {
		t.Errorf("client owner: create = %d, want 200", code)
	}
}
