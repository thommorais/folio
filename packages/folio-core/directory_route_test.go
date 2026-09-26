package folio_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	folio "folio/folio-core"
	"folio/folio-core/adapters/httpapi"
	"folio/folio-core/adapters/pb"
)

func TestClientsAndDomainsOverTheAPI(t *testing.T) {
	dir, err := os.MkdirTemp("", "folio-directory")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: dir})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.ResetBootstrapState() })
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	if err := folio.Migrate(app); err != nil {
		t.Fatal(err)
	}

	owner := record(t, app, pb.ColUsers, map[string]any{"email": "owner@test.local", "password": "password12345", "verified": true})
	client := record(t, app, pb.ColClients, map[string]any{"slug": "acme", "name": "Acme", "site": "https://acme.test"})
	record(t, app, pb.ColClientMembers, map[string]any{"client": client.Id, "user": owner.Id, "role": "owner"})
	dom := record(t, app, pb.ColDomains, map[string]any{"client": client.Id, "slug": "web", "name": "Web"})
	record(t, app, pb.ColMembers, map[string]any{"domain": dom.Id, "project": record(t, app, pb.ColProjects, map[string]any{
		"slug": "redesign", "name": "Redesign", "domain": dom.Id,
	}).Id, "user": owner.Id, "role": "owner"})

	router, err := apis.NewRouter(app)
	if err != nil {
		t.Fatal(err)
	}
	httpapi.New(folio.New(app, nil).Deps()).Mount(&core.ServeEvent{App: app, Router: router})
	mux, err := router.BuildMux()
	if err != nil {
		t.Fatal(err)
	}
	token, err := owner.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	get := func(path string, out any) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s = %d %s", path, rec.Code, rec.Body)
		}
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatal(err)
		}
	}

	var clients struct {
		Clients []map[string]any `json:"clients"`
	}
	get("/api/folio/clients", &clients)
	if len(clients.Clients) != 1 || clients.Clients[0]["slug"] != "acme" || clients.Clients[0]["site"] != "https://acme.test" {
		t.Errorf("clients = %v, want acme with its site", clients.Clients)
	}

	var domains struct {
		Domains []map[string]any `json:"domains"`
	}
	get("/api/folio/domains", &domains)
	if len(domains.Domains) != 1 {
		t.Fatalf("domains = %v, want one", domains.Domains)
	}
	d := domains.Domains[0]
	if d["slug"] != "web" || d["client_slug"] != "acme" || d["client_id"] != client.Id || d["id"] != dom.Id {
		t.Errorf("domain = %v, want web under acme", d)
	}
	if members, _ := d["members"].([]any); len(members) != 1 {
		t.Errorf("domain members = %v, want the owner", d["members"])
	}
}
