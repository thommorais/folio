package main

import (
	"net/http"
	"os"
	"testing"
)

const aDomain = `{"id":"d1","client_id":"c1","client_slug":"acme","slug":"web","name":"Web","members":[{"user_id":"u1","role":"owner"}],"created_at":"2026-09-26T10:00:00Z","updated_at":"2026-09-26T10:00:00Z"}`

func runDomain(t *testing.T, args ...string) error {
	t.Helper()

	stdout := os.Stdout
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	os.Stdout = devnull
	t.Cleanup(func() {
		os.Stdout = stdout
		_ = devnull.Close()
	})

	cmd := domainCommand()
	cmd.SetOut(devnull)
	cmd.SetErr(devnull)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestDomainListReadsTheDirectory(t *testing.T) {
	got := knowledgeServer(t, `{"domains":[`+aDomain+`]}`)

	if err := runDomain(t, "list"); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 || (*got)[0].method != http.MethodGet || (*got)[0].path != "/api/folio/domains" {
		t.Fatalf("requests = %+v, want GET /api/folio/domains", *got)
	}
}

func TestDomainUpdateAddressesTheDomainUnderItsClient(t *testing.T) {
	got := knowledgeServer(t, aDomain)

	if err := runDomain(t, "update", "acme/web", "--name", "Web Team"); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Fatalf("requests = %+v, want one", *got)
	}
	req := (*got)[0]
	if req.method != http.MethodPatch || req.path != "/api/folio/clients/acme/domains/web" {
		t.Errorf("%s %s, want PATCH /api/folio/clients/acme/domains/web", req.method, req.path)
	}
	if req.body["name"] != "Web Team" || len(req.body) != 1 {
		t.Errorf("body = %v, want only the name", req.body)
	}
}

func TestDomainUpdateRejectsABadReference(t *testing.T) {
	for _, ref := range []string{"web", "acme/", "/web", "acme/web/extra"} {
		t.Run(ref, func(t *testing.T) {
			got := knowledgeServer(t, aDomain)
			if err := runDomain(t, "update", ref, "--name", "X"); err == nil {
				t.Fatalf("update %q succeeded", ref)
			}
			if len(*got) != 0 {
				t.Errorf("sent %d requests for a bad reference", len(*got))
			}
		})
	}
}

func TestDomainUpdateRefusesAnEmptyChange(t *testing.T) {
	got := knowledgeServer(t, aDomain)

	if err := runDomain(t, "update", "acme/web"); err == nil {
		t.Fatal("update with no fields succeeded")
	}
	if len(*got) != 0 {
		t.Errorf("sent %d requests for an empty update", len(*got))
	}
}
