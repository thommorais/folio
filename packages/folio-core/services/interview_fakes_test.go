package services_test

import (
	"context"
	"sort"
	"time"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

type fakeInterviews struct {
	items   map[domain.InterviewID]domain.Interview
	events  map[domain.InterviewID][]domain.InterviewEvent
	issues  *fakeIssues
	entries *fakeEntries
}

func newFakeInterviews(issues *fakeIssues, entries *fakeEntries) *fakeInterviews {
	return &fakeInterviews{
		items: map[domain.InterviewID]domain.Interview{}, events: map[domain.InterviewID][]domain.InterviewEvent{},
		issues: issues, entries: entries,
	}
}

var _ ports.InterviewRepository = (*fakeInterviews)(nil)

func (r *fakeInterviews) Create(_ context.Context, i domain.Interview) (domain.Interview, error) {
	r.items[i.ID] = i
	return i, nil
}

func (r *fakeInterviews) GetByID(_ context.Context, id domain.InterviewID) (domain.Interview, error) {
	i, ok := r.items[id]
	if !ok {
		return domain.Interview{}, domain.ErrNotFound
	}
	return i, nil
}

func (r *fakeInterviews) ListByIssue(_ context.Context, issue domain.IssueID) ([]domain.Interview, error) {
	out := []domain.Interview{}
	for _, i := range r.items {
		if i.IssueID == issue {
			out = append(out, i)
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].ID < out[b].ID })
	return out, nil
}

func (r *fakeInterviews) Update(_ context.Context, i domain.Interview) (domain.Interview, error) {
	if _, ok := r.items[i.ID]; !ok {
		return domain.Interview{}, domain.ErrNotFound
	}
	r.items[i.ID] = i
	return i, nil
}

func (r *fakeInterviews) AppendEvent(_ context.Context, i domain.Interview, actions []domain.SendAction, at time.Time) (domain.InterviewEvent, error) {
	event := domain.InterviewEvent{
		ID: domain.EventID(string(i.ID) + "-" + string(rune('0'+len(r.events[i.ID])+1))), InterviewID: i.ID,
		Seq: len(r.events[i.ID]) + 1, At: at, Actions: actions,
	}
	r.events[i.ID] = append(r.events[i.ID], event)
	return event, nil
}

func (r *fakeInterviews) EventsAfter(_ context.Context, id domain.InterviewID, seq int) ([]domain.InterviewEvent, error) {
	out := []domain.InterviewEvent{}
	for _, e := range r.events[id] {
		if e.Seq > seq {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *fakeInterviews) Finish(ctx context.Context, i domain.Interview, ticket domain.Issue, resolution domain.Entry) error {
	if _, err := r.entries.Create(ctx, resolution); err != nil {
		return err
	}
	if _, err := r.issues.Update(ctx, ticket); err != nil {
		return err
	}
	_, err := r.Update(ctx, i)
	return err
}
