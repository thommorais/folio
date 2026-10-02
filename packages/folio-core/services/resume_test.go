package services_test

import (
	"testing"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

func TestResumeCarriesTheLatestHandoffAndWhatCameAfter(t *testing.T) {
	f := newEntryFixture(t)
	ctx := t.Context()
	work, err := f.issueSv.CreateIssue(ctx, f.owner, ports.CreateIssueInput{ProjectID: f.project, Kind: domain.IssueTicket, Title: "Landing page"})
	if err != nil {
		t.Fatal(err)
	}

	at := func(minutes int) time.Time { return testNow.Add(time.Duration(minutes) * time.Minute) }
	for _, e := range []domain.Entry{
		{ID: "log-before", Kind: domain.EntryLog, Body: "before", CreatedAt: at(1)},
		{ID: "handoff-new", Kind: domain.EntryHandoff, Body: "Next: coverage widget.", Meta: map[string]any{"commit": "abc"}, CreatedAt: at(3)},
		{ID: "handoff-old", Kind: domain.EntryHandoff, Body: "Next: get started.", CreatedAt: at(2)},
		{ID: "log-after", Kind: domain.EntryLog, Body: "after", CreatedAt: at(4)},
		{ID: "journal", Kind: domain.EntryJournal, Slug: "j", Title: "J", CreatedAt: at(5)},
	} {
		e.ProjectID, e.IssueID = f.project, work.ID
		if _, err := f.entries.Create(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	resume, err := f.issueSv.GetIssueResume(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resume.Issue.ID != work.ID {
		t.Errorf("issue = %s, want %s", resume.Issue.ID, work.ID)
	}
	if resume.Handoff == nil || resume.Handoff.ID != "handoff-new" {
		t.Fatalf("handoff = %+v, want the latest one", resume.Handoff)
	}
	if len(resume.Logs) != 1 || resume.Logs[0].ID != "log-after" {
		t.Errorf("logs = %+v, want only the log written after the handoff", resume.Logs)
	}
}

func TestResumeWithoutAHandoffCarriesTheRecentLogs(t *testing.T) {
	f := newEntryFixture(t)
	ctx := t.Context()
	work, err := f.issueSv.CreateIssue(ctx, f.owner, ports.CreateIssueInput{ProjectID: f.project, Kind: domain.IssueTicket, Title: "Landing page"})
	if err != nil {
		t.Fatal(err)
	}
	for i := range 8 {
		e := domain.Entry{
			ID: domain.EntryID("log" + string(rune('a'+i))), Kind: domain.EntryLog, ProjectID: f.project, IssueID: work.ID,
			Body: "step", CreatedAt: testNow.Add(time.Duration(8-i) * time.Minute),
		}
		if _, err := f.entries.Create(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	resume, err := f.issueSv.GetIssueResume(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resume.Handoff != nil {
		t.Fatalf("handoff = %+v, want none", resume.Handoff)
	}
	if len(resume.Logs) != 5 {
		t.Fatalf("logs = %d, want the 5 most recent", len(resume.Logs))
	}
	if resume.Logs[0].ID != "loge" || resume.Logs[4].ID != "loga" {
		t.Errorf("logs = %s..%s, want oldest to newest of the last five", resume.Logs[0].ID, resume.Logs[4].ID)
	}
}

func TestResumeListsOpenChildrenAndCountsTheRest(t *testing.T) {
	f := newEntryFixture(t)
	ctx := t.Context()
	work, err := f.issueSv.CreateIssue(ctx, f.owner, ports.CreateIssueInput{ProjectID: f.project, Kind: domain.IssueTicket, Title: "Landing page"})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		title  string
		status domain.IssueStatus
	}{{"Coverage", domain.IssueOpen}, {"Icons", domain.IssueDone}, {"Palette", domain.IssueInProgress}, {"Dropped", domain.IssueCancelled}} {
		child, err := f.issueSv.CreateIssue(ctx, f.owner, ports.CreateIssueInput{ProjectID: f.project, Kind: domain.IssueTodo, Title: c.title, ParentID: work.ID})
		if err != nil {
			t.Fatal(err)
		}
		if c.status != domain.IssueOpen {
			if _, err := f.issueSv.UpdateIssue(ctx, f.owner, child.ID, ports.UpdateIssueInput{Status: ptr(c.status)}); err != nil {
				t.Fatal(err)
			}
		}
	}

	resume, err := f.issueSv.GetIssueResume(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	titles := []string{}
	for _, c := range resume.Open {
		titles = append(titles, c.Title)
	}
	if len(titles) != 2 || titles[0] != "Coverage" || titles[1] != "Palette" {
		t.Errorf("open = %v, want Coverage and Palette", titles)
	}
	if resume.Closed != 2 {
		t.Errorf("closed = %d, want 2", resume.Closed)
	}
}

func TestResumeIsFencedByMembership(t *testing.T) {
	f := newEntryFixture(t)
	work, err := f.issueSv.CreateIssue(t.Context(), f.owner, ports.CreateIssueInput{ProjectID: f.project, Kind: domain.IssueTicket, Title: "Landing page"})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.issueSv.GetIssueResume(t.Context(), f.outside, work.ID); err == nil {
		t.Fatal("want a stranger refused")
	}
}

func TestResumeLeavesArchivedChildrenOut(t *testing.T) {
	f := newEntryFixture(t)
	ctx := t.Context()
	work, err := f.issueSv.CreateIssue(ctx, f.owner, ports.CreateIssueInput{ProjectID: f.project, Kind: domain.IssueTicket, Title: "Landing page"})
	if err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{"Coverage", "Shelved"} {
		child, err := f.issueSv.CreateIssue(ctx, f.owner, ports.CreateIssueInput{ProjectID: f.project, Kind: domain.IssueTodo, Title: title, ParentID: work.ID})
		if err != nil {
			t.Fatal(err)
		}
		if title == "Shelved" {
			if _, err := f.issueSv.UpdateIssue(ctx, f.owner, child.ID, ports.UpdateIssueInput{Archived: ptr(true)}); err != nil {
				t.Fatal(err)
			}
		}
	}

	resume, err := f.issueSv.GetIssueResume(ctx, f.owner, work.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(resume.Open) != 1 || resume.Open[0].Title != "Coverage" {
		t.Errorf("open = %v, want only Coverage", resume.Open)
	}
}
