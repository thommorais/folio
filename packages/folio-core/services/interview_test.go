package services_test

import (
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

const firstRound = `{"questions":[
	{"id":"q1","round":1,"title":"Tree or graph","options":[{"k":"a","text":"Tree"},{"k":"b","text":"Graph"}],"rec":{"option":"b","why":"Edges show blockers."}},
	{"id":"q2","round":1,"title":"What do we call it","rec":{"text":"interview","why":"Plain word."}}
]}`

func (f *ticketFixture) grilling(t *testing.T, title string) domain.Issue {
	t.Helper()
	return f.decision(t, domain.WayfinderGrilling, title)
}

func TestStartResumesTheActiveInterview(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	ticket := f.grilling(t, "Tree or graph")

	first, created, err := f.interviewSvc.StartInterview(ctx, f.owner, ticket.ID)
	if err != nil || !created {
		t.Fatalf("start = %v, created %v", err, created)
	}
	if first.Topic != "Tree or graph" || first.AgentStatus != domain.AgentWaiting || first.IssueID != ticket.ID {
		t.Fatalf("interview = %+v", first)
	}
	again, created, err := f.interviewSvc.StartInterview(ctx, f.owner, ticket.ID)
	if err != nil || created || again.ID != first.ID {
		t.Fatalf("a second start must resume %s, got %s created %v err %v", first.ID, again.ID, created, err)
	}
}

func TestStartRefusesWhatIsNotAnOpenGrillingTicket(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()

	if _, _, err := f.interviewSvc.StartInterview(ctx, f.owner, f.ticket(t, f.project, "Mobile nav").ID); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("a plain ticket: want validation error, got %v", err)
	}
	if _, _, err := f.interviewSvc.StartInterview(ctx, f.owner, f.decision(t, domain.WayfinderResearch, "How dense").ID); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("a research ticket: want validation error, got %v", err)
	}
	closed := f.grilling(t, "Closed already")
	if _, err := f.issueSvc.UpdateIssue(ctx, f.owner, closed.ID, ports.UpdateIssueInput{Status: ptr(domain.IssueDone), Resolution: ptr("A graph.")}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.interviewSvc.StartInterview(ctx, f.owner, closed.ID); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("a closed ticket: want validation error, got %v", err)
	}
	if _, _, err := f.interviewSvc.StartInterview(ctx, f.viewer, f.grilling(t, "Viewer tries").ID); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("a viewer: want forbidden, got %v", err)
	}
}

func TestTheTurnLoop(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	ticket := f.grilling(t, "Tree or graph")
	if _, _, err := f.interviewSvc.StartInterview(ctx, f.owner, ticket.ID); err != nil {
		t.Fatal(err)
	}

	summary, err := f.interviewSvc.PatchInterview(ctx, f.owner, ticket.ID, []byte(firstRound))
	if err != nil {
		t.Fatal(err)
	}
	if summary.Round != 1 || summary.Added != 2 || summary.Answered != 0 {
		t.Fatalf("summary = %+v", summary)
	}

	if pending, err := f.interviewSvc.PendingSends(ctx, f.owner, ticket.ID); err != nil || len(pending) != 0 {
		t.Fatalf("nothing sent yet: %v, %v", pending, err)
	}

	send := []domain.SendAction{
		{Type: domain.SendAnswer, Q: "q1", Kind: domain.AnswerAccept, Option: "b"},
		{Type: domain.SendThread, Q: "q2", Text: "why not grill?"},
	}
	event, err := f.interviewSvc.SendToInterview(ctx, f.owner, ticket.ID, send)
	if err != nil || event.Seq != 1 {
		t.Fatalf("send = %+v, %v", event, err)
	}
	if _, err := f.interviewSvc.SendToInterview(ctx, f.owner, ticket.ID, []domain.SendAction{{Type: domain.SendDefer, Q: "q9"}}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("a Send naming no question: want validation error, got %v", err)
	}

	pending, err := f.interviewSvc.PendingSends(ctx, f.owner, ticket.ID)
	if err != nil || len(pending) != 1 || pending[0].Seq != 1 {
		t.Fatalf("pending = %+v, %v", pending, err)
	}
	working, err := f.interviewSvc.CurrentInterview(ctx, f.owner, ticket.ID)
	if err != nil || working.AgentStatus != domain.AgentWorking {
		t.Fatalf("reading pending Sends marks the agent working, got %s, %v", working.AgentStatus, err)
	}

	if _, err := f.interviewSvc.PatchInterview(ctx, f.owner, ticket.ID, []byte(`{"agent":{"handled":2}}`)); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("handled past the last Send: want validation error, got %v", err)
	}
	reply := `{"questions":[{"id":"q1","status":"answered","answer":{"kind":"accept","option":"b"}},{"id":"q2","thread":[{"who":"agent","text":"grill is the skill's verb."}]}],"agent":{"handled":1}}`
	summary, err = f.interviewSvc.PatchInterview(ctx, f.owner, ticket.ID, []byte(reply))
	if err != nil {
		t.Fatal(err)
	}
	if summary.Answered != 1 || summary.Handled != 1 || summary.Interview.AgentStatus != domain.AgentWaiting {
		t.Fatalf("summary = %+v", summary)
	}
	if pending, _ := f.interviewSvc.PendingSends(ctx, f.owner, ticket.ID); len(pending) != 0 {
		t.Fatalf("everything is handled, got %+v", pending)
	}
}

func TestFinishResolvesTheTicket(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	ticket := f.grilling(t, "Tree or graph")
	if _, _, err := f.interviewSvc.StartInterview(ctx, f.owner, ticket.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.interviewSvc.PatchInterview(ctx, f.owner, ticket.ID, []byte(firstRound)); err != nil {
		t.Fatal(err)
	}

	if _, err := f.interviewSvc.FinishInterview(ctx, f.owner, ticket.ID, "A graph.", "# Doc"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("open questions: want validation error, got %v", err)
	}

	settle := `{"questions":[{"id":"q1","status":"answered","answer":{"kind":"accept","option":"b"}},{"id":"q2","status":"deferred"}]}`
	if _, err := f.interviewSvc.PatchInterview(ctx, f.owner, ticket.ID, []byte(settle)); err != nil {
		t.Fatal(err)
	}
	finished, err := f.interviewSvc.FinishInterview(ctx, f.owner, ticket.ID, "A graph.", "# Tree or graph\n\nA graph.")
	if err != nil {
		t.Fatal(err)
	}
	if !finished.IsFinished() {
		t.Fatal("the interview should be finished")
	}

	closed, err := f.issueSvc.GetIssue(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if closed.Status != domain.IssueDone || closed.Resolution != "A graph." || closed.ResolutionEntry == "" {
		t.Fatalf("ticket = %+v", closed)
	}
	detail, err := f.entries.GetByID(ctx, closed.ResolutionEntry)
	if err != nil || detail.Kind != domain.EntryResolution || detail.IssueID != ticket.ID || detail.Body != "# Tree or graph\n\nA graph." {
		t.Fatalf("detail = %+v, %v", detail, err)
	}

	if _, err := f.interviewSvc.PatchInterview(ctx, f.owner, ticket.ID, []byte(`{"note":"late"}`)); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a finished interview is not active any more, got %v", err)
	}
	if list, _ := f.interviewSvc.ListInterviews(ctx, f.owner, ticket.ID); len(list) != 1 || !list[0].IsFinished() {
		t.Fatalf("the finished interview stays listed, got %+v", list)
	}
}

func TestFinishMarksTheFinishSendHandled(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	ticket := f.grilling(t, "Tree or graph")
	if _, _, err := f.interviewSvc.StartInterview(ctx, f.owner, ticket.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.interviewSvc.PatchInterview(ctx, f.owner, ticket.ID, []byte(firstRound)); err != nil {
		t.Fatal(err)
	}
	settle := `{"questions":[{"id":"q1","status":"answered","answer":{"kind":"accept","option":"b"}},{"id":"q2","status":"deferred"}]}`
	if _, err := f.interviewSvc.PatchInterview(ctx, f.owner, ticket.ID, []byte(settle)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.interviewSvc.SendToInterview(ctx, f.owner, ticket.ID, []domain.SendAction{{Type: domain.SendFinish}}); err != nil {
		t.Fatal(err)
	}

	finished, err := f.interviewSvc.FinishInterview(ctx, f.owner, ticket.ID, "A graph.", "# Doc")
	if err != nil {
		t.Fatal(err)
	}
	if finished.Handled != 1 {
		t.Fatalf("handled = %d, want 1: the finish Send is consumed by finishing", finished.Handled)
	}
}

func TestPagePathNamesTheTicketByItsSlugs(t *testing.T) {
	f := newTicketFixture(t)
	ctx := t.Context()
	ticket := f.grilling(t, "Tree or graph")

	got, err := f.interviewSvc.PagePath(ctx, f.owner, ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "/acme/web/api/tickets/" + ticket.Slug + "/interview"; got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestPagePathIsEmptyForAProjectWithoutADomain(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.ticket(t, f.other, "Tree or graph")

	got, err := f.interviewSvc.PagePath(t.Context(), f.owner, ticket.ID)
	if err != nil || got != "" {
		t.Fatalf("path = %q, %v, want empty", got, err)
	}
}

func TestPagePathNeedsReadAccess(t *testing.T) {
	f := newTicketFixture(t)
	ticket := f.grilling(t, "Tree or graph")

	if _, err := f.interviewSvc.PagePath(t.Context(), f.outside, ticket.ID); !errors.Is(err, domain.ErrNotFound) && !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("a stranger should be refused, got %v", err)
	}
}
