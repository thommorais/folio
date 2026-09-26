package main

import (
	"net/http"
	"os"
	"testing"
)

const aClient = `{"id":"c1","slug":"acme","name":"Acme","site":"https://acme.test","created_at":"2026-09-26T10:00:00Z","updated_at":"2026-09-26T10:00:00Z"}`

func runClient(t *testing.T, args ...string) error {
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

	cmd := clientCommand()
	cmd.SetOut(devnull)
	cmd.SetErr(devnull)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestClientListReadsTheDirectory(t *testing.T) {
	got := knowledgeServer(t, `{"clients":[`+aClient+`]}`)

	if err := runClient(t, "list"); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 || (*got)[0].method != http.MethodGet || (*got)[0].path != "/api/folio/clients" {
		t.Fatalf("requests = %+v, want GET /api/folio/clients", *got)
	}
}

func TestClientUpdateSendsOnlyGivenFields(t *testing.T) {
	got := knowledgeServer(t, aClient)

	if err := runClient(t, "update", "acme", "--name", "Acme Corp", "--site", "https://acme.example"); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Fatalf("requests = %+v, want one", *got)
	}
	req := (*got)[0]
	if req.method != http.MethodPatch || req.path != "/api/folio/clients/acme" {
		t.Errorf("%s %s, want PATCH /api/folio/clients/acme", req.method, req.path)
	}
	if req.body["name"] != "Acme Corp" || req.body["site"] != "https://acme.example" || len(req.body) != 2 {
		t.Errorf("body = %v, want only name and site", req.body)
	}
}

func TestClientUpdateRefusesAnEmptyChange(t *testing.T) {
	got := knowledgeServer(t, aClient)

	if err := runClient(t, "update", "acme"); err == nil {
		t.Fatal("update with no fields succeeded")
	}
	if len(*got) != 0 {
		t.Errorf("sent %d requests for an empty update", len(*got))
	}
}
