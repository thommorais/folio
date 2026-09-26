package pb_test

import (
	"testing"
	"time"

	"folio/folio-core/adapters/pb"
	"folio/folio-core/domain"
)

func anInterview(s scenario, issue domain.IssueID) domain.Interview {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	return domain.Interview{
		ProjectID: domain.ProjectID(s.project.Id), IssueID: issue, Topic: "Tree or graph",
		State: domain.InterviewState{
			Terms: []domain.Term{{Term: "interview", Def: "one asking", Avoid: []string{"grill"}}},
			Questions: []domain.Question{{
				ID: "q1", Round: 1, Title: "Tree or graph", Status: domain.QuestionOpen, Deps: []string{},
				Options: []domain.Option{{Key: "a", Text: "Tree"}, {Key: "b", Text: "Graph"}},
				Rec:     domain.Recommendation{Option: "b", Why: "Edges show blockers."},
				Thread:  []domain.Message{{Who: domain.SpeakerUser, Text: "why?", At: now}},
			}},
		},
		AgentStatus: domain.AgentWaiting, AgentSince: now,
	}
}

func TestInterviewRoundTrip(t *testing.T) {
	s := setup(t)
	repo := pb.NewInterviewRepository(s.app)
	ticket := newIssue(t, s, domain.IssueTicket, "tree-or-graph")

	created, err := repo.Create(t.Context(), anInterview(s, ticket.ID))
	if err != nil {
		t.Fatal(err)
	}
	created.Handled = 4
	created.State.Note = "one branch left"
	if _, err := repo.Update(t.Context(), created); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Topic != "Tree or graph" || got.Handled != 4 || got.AgentStatus != domain.AgentWaiting || got.IsFinished() {
		t.Fatalf("interview = %+v", got)
	}
	if got.State.Note != "one branch left" || len(got.State.Questions) != 1 || got.State.Questions[0].Rec.Option != "b" ||
		len(got.State.Questions[0].Thread) != 1 || got.State.Terms[0].Avoid[0] != "grill" {
		t.Fatalf("state = %+v", got.State)
	}
}

func TestInterviewsListOldestFirst(t *testing.T) {
	s := setup(t)
	repo := pb.NewInterviewRepository(s.app)
	ticket := newIssue(t, s, domain.IssueTicket, "tree-or-graph")

	first, err := repo.Create(t.Context(), anInterview(s, ticket.ID))
	if err != nil {
		t.Fatal(err)
	}
	done := time.Date(2026, 9, 26, 13, 0, 0, 0, time.UTC)
	first.FinishedAt = &done
	if _, err := repo.Update(t.Context(), first); err != nil {
		t.Fatal(err)
	}
	second, err := repo.Create(t.Context(), anInterview(s, ticket.ID))
	if err != nil {
		t.Fatal(err)
	}

	list, err := repo.ListByIssue(t.Context(), ticket.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ID != first.ID || list[1].ID != second.ID || !list[0].IsFinished() {
		t.Fatalf("list = %+v", list)
	}
}

func TestEventsNumberPerInterview(t *testing.T) {
	s := setup(t)
	repo := pb.NewInterviewRepository(s.app)
	ticket := newIssue(t, s, domain.IssueTicket, "tree-or-graph")
	one, _ := repo.Create(t.Context(), anInterview(s, ticket.ID))
	two, _ := repo.Create(t.Context(), anInterview(s, ticket.ID))
	at := time.Date(2026, 9, 26, 12, 30, 0, 0, time.UTC)
	send := []domain.SendAction{{Type: domain.SendAnswer, Q: "q1", Kind: domain.AnswerAccept, Option: "b"}}

	for i := 0; i < 3; i++ {
		event, err := repo.AppendEvent(t.Context(), one, send, at)
		if err != nil {
			t.Fatal(err)
		}
		if event.Seq != i+1 {
			t.Fatalf("seq = %d, want %d", event.Seq, i+1)
		}
	}
	other, err := repo.AppendEvent(t.Context(), two, send, at)
	if err != nil || other.Seq != 1 {
		t.Fatalf("another interview starts its own count, got %d, %v", other.Seq, err)
	}

	after, err := repo.EventsAfter(t.Context(), one.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 2 || after[0].Seq != 2 || after[1].Seq != 3 || after[0].Actions[0].Option != "b" || !after[0].At.Equal(at) {
		t.Fatalf("after 1 = %+v", after)
	}
}

func TestInterviewsGoWithTheirTicket(t *testing.T) {
	s := setup(t)
	repo := pb.NewInterviewRepository(s.app)
	ticket := newIssue(t, s, domain.IssueTicket, "tree-or-graph")
	interview, err := repo.Create(t.Context(), anInterview(s, ticket.ID))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AppendEvent(t.Context(), interview, []domain.SendAction{{Type: domain.SendFinish}}, time.Now()); err != nil {
		t.Fatal(err)
	}

	if err := pb.NewIssueRepository(s.app).Delete(t.Context(), ticket.ID); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.ListByIssue(t.Context(), ticket.ID); len(list) != 0 {
		t.Fatalf("interviews left behind: %+v", list)
	}
	if n, _ := s.app.CountRecords(pb.ColInterviewEvents); n != 0 {
		t.Fatalf("events left behind: %d", n)
	}
}
