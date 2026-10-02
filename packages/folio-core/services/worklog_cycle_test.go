package services_test

import (
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func TestWorkLogTakesTheOpenCycle(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	ticket := f.ticket(t, f.project, "Mobile nav")

	write := func(t *testing.T, kind domain.EntryKind, body string) domain.Entry {
		t.Helper()
		in := ports.WriteEntryInput{ProjectID: f.project, Kind: kind, IssueID: ticket.ID, Body: body}
		if kind == domain.EntryJournal {
			in.Title = body
		}
		got, err := f.entrySvc.WriteEntry(ctx, f.owner, in)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}

	if got := write(t, domain.EntryLog, "Started looking"); got.CycleID != "" {
		t.Errorf("no cycle open: cycle = %q, want empty", got.CycleID)
	}

	cycle, err := f.cycleSvc.OpenCycle(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := write(t, domain.EntryLog, "Mapbox rejects feature-state in a filter"); got.CycleID != cycle.ID {
		t.Errorf("cycle open: cycle = %q, want %q", got.CycleID, cycle.ID)
	}
	if got := write(t, domain.EntryJournal, "A journal entry"); got.CycleID != "" {
		t.Errorf("a journal entry is not a work log: cycle = %q, want empty", got.CycleID)
	}

	if _, err := f.cycleSvc.ResolveCycle(ctx, f.owner, cycle.ID, "Shipped"); err != nil {
		t.Fatal(err)
	}
	if got := write(t, domain.EntryLog, "After the cycle"); got.CycleID != "" {
		t.Errorf("cycle resolved: cycle = %q, want empty", got.CycleID)
	}
}
