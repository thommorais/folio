package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const aProject = `{"id":"p1","domain_id":"d1","slug":"site","name":"Site","archived":false,"members":[],"created_at":"2026-09-26T10:00:00Z","updated_at":"2026-09-26T10:00:00Z"}`

func projectServer(t *testing.T) *[]request {
	t.Helper()

	var got []request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := request{method: r.Method, path: r.URL.Path}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &req.body)
		}
		got = append(got, req)
		if r.URL.Path == "/api/folio/domains" {
			_, _ = w.Write([]byte(`{"domains":[` + aDomain + `]}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(aProject))
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")
	return &got
}

func runProject(t *testing.T, args ...string) error {
	t.Helper()

	stdout, stderr := os.Stdout, os.Stderr
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	os.Stdout, os.Stderr = devnull, devnull
	t.Cleanup(func() {
		os.Stdout, os.Stderr = stdout, stderr
		_ = devnull.Close()
	})

	cmd := projectCommand()
	cmd.SetOut(devnull)
	cmd.SetErr(devnull)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestProjectCreateResolvesTheDomain(t *testing.T) {
	for _, ref := range []string{"acme/web", "c1/d1", "acme/d1"} {
		t.Run(ref, func(t *testing.T) {
			got := projectServer(t)

			if err := runProject(t, "create", "Site", "--domain", ref); err != nil {
				t.Fatal(err)
			}
			if len(*got) != 2 || (*got)[0].path != "/api/folio/domains" {
				t.Fatalf("requests = %+v, want the domain lookup then the create", *got)
			}
			create := (*got)[1]
			if create.method != http.MethodPost || create.path != "/api/folio/projects" || create.body["domain_id"] != "d1" {
				t.Errorf("create = %+v, want POST /api/folio/projects with domain_id d1", create)
			}
		})
	}
}

func TestProjectCreateWithoutDomainSkipsTheLookup(t *testing.T) {
	got := projectServer(t)

	if err := runProject(t, "create", "Site"); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 || (*got)[0].method != http.MethodPost {
		t.Fatalf("requests = %+v, want only the create", *got)
	}
	if _, sent := (*got)[0].body["domain_id"]; sent {
		t.Error("domain_id was sent without --domain")
	}
}

func TestProjectCreateStopsOnAnUnknownDomain(t *testing.T) {
	for _, ref := range []string{"acme/mobile", "globex/web"} {
		t.Run(ref, func(t *testing.T) {
			got := projectServer(t)

			if err := runProject(t, "create", "Site", "--domain", ref); err == nil {
				t.Fatal("create into an unknown domain succeeded")
			}
			for _, req := range *got {
				if req.method == http.MethodPost {
					t.Errorf("a project was created: %+v", req)
				}
			}
		})
	}
}

func TestProjectCreateRejectsABadDomainReference(t *testing.T) {
	got := projectServer(t)

	if err := runProject(t, "create", "Site", "--domain", "web"); err == nil {
		t.Fatal("create with a bad domain reference succeeded")
	}
	if len(*got) != 0 {
		t.Errorf("sent %d requests for a bad reference", len(*got))
	}
}
