package main

import (
	"io"
	"os"
	"testing"

	"folio/cli/internal/client"
)

func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = stdout

	printed, _ := io.ReadAll(r)
	_ = r.Close()
	if runErr != nil {
		t.Fatal(runErr)
	}
	return string(printed)
}

func TestTicketDetailKeepsAFixedFieldOrder(t *testing.T) {
	ticket := client.Ticket{
		ID: "t1", Title: "Nav", Status: "open", Priority: "medium",
		Assignee: "u1", ExternalRef: "JIRA-1", ParentID: "p1", Wayfinder: "map",
	}
	want := "Nav\nt1  open  medium  0/0 done\n" +
		"assignee: u1\nexternal ref: JIRA-1\nparent: p1\nwayfinder: map\n"

	for range 30 {
		got := captureStdout(t, func() error { return renderTicketDetail(ticket) })
		if got != want {
			t.Fatalf("output =\n%q\nwant\n%q", got, want)
		}
	}
}

func TestLogDetailKeepsAFixedFieldOrder(t *testing.T) {
	entry := client.JournalEntry{
		Title: "Shipped", Slug: "shipped", Branch: "develop", PR: "https://x/pull/1", ExternalRef: "JIRA-1",
	}
	want := "Shipped\n=======\nslug: shipped\nbranch: develop\npr: https://x/pull/1\nexternal ref: JIRA-1\n"

	for range 30 {
		got := captureStdout(t, func() error { return renderLogDetail(entry) })
		if got != want {
			t.Fatalf("output =\n%q\nwant\n%q", got, want)
		}
	}
}
