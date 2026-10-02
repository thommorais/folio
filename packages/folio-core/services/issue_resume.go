package services

import (
	"context"
	"sort"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

const RecentLogsWithoutHandoff = 5

func (s *IssueService) GetIssueResume(ctx context.Context, actor ports.Actor, id domain.IssueID) (domain.IssueResume, error) {
	issue, err := s.GetIssue(ctx, actor, id)
	if err != nil {
		return domain.IssueResume{}, err
	}
	return s.resume(ctx, issue)
}

func (s *IssueService) GetIssueResumeBySlug(ctx context.Context, actor ports.Actor, project domain.ProjectID, slug string) (domain.IssueResume, error) {
	issue, err := s.GetIssueBySlug(ctx, actor, project, slug)
	if err != nil {
		return domain.IssueResume{}, err
	}
	return s.resume(ctx, issue)
}

func (s *IssueService) resume(ctx context.Context, issue domain.Issue) (domain.IssueResume, error) {
	out := domain.IssueResume{Issue: issue, Open: []domain.Issue{}, Plans: []domain.Plan{}}

	handoffs, err := s.entries.List(ctx, issue.ProjectID, domain.EntryFilter{Kind: domain.EntryHandoff, IssueID: issue.ID, Limit: MaxPageSize})
	if err != nil {
		return domain.IssueResume{}, err
	}
	sortOldestFirst(handoffs)
	if len(handoffs) > 0 {
		out.Handoff = &handoffs[len(handoffs)-1]
	}

	logs, err := s.entries.List(ctx, issue.ProjectID, domain.EntryFilter{Kind: domain.EntryLog, IssueID: issue.ID, Limit: MaxPageSize})
	if err != nil {
		return domain.IssueResume{}, err
	}
	sortOldestFirst(logs)
	out.Logs = logsToResume(logs, out.Handoff)

	children, err := s.repo.ListByParent(ctx, issue.ID)
	if err != nil {
		return domain.IssueResume{}, err
	}
	if err := s.decorate(ctx, issue.ProjectID, children); err != nil {
		return domain.IssueResume{}, err
	}
	for _, c := range children {
		if c.Status.IsTerminal() {
			out.Closed++
			continue
		}
		out.Open = append(out.Open, c)
	}

	plans, err := s.plans.ListByIssue(ctx, issue.ID)
	if err != nil {
		return domain.IssueResume{}, err
	}
	for _, p := range plans {
		if !p.Status.IsTerminal() {
			out.Plans = append(out.Plans, p)
		}
	}

	cycles, err := s.cycles.ListByIssue(ctx, issue.ID)
	if err != nil {
		return domain.IssueResume{}, err
	}
	if current, ok := rules.CurrentCycle(cycles); ok {
		out.Cycle = &current
	}
	if out.Map, err = s.mapBrief(ctx, issue.ProjectID, cycles); err != nil {
		return domain.IssueResume{}, err
	}

	if out.Docs, err = s.entries.List(ctx, issue.ProjectID, domain.EntryFilter{Kind: domain.EntryDoc, IssueID: issue.ID, Limit: MaxPageSize}); err != nil {
		return domain.IssueResume{}, err
	}
	return out, nil
}

func logsToResume(logs []domain.Entry, handoff *domain.Entry) []domain.Entry {
	if handoff == nil {
		return logs[max(0, len(logs)-RecentLogsWithoutHandoff):]
	}
	after := []domain.Entry{}
	for _, l := range logs {
		if l.CreatedAt.After(handoff.CreatedAt) {
			after = append(after, l)
		}
	}
	return after
}

func sortOldestFirst(entries []domain.Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].CreatedAt.Before(entries[j].CreatedAt)
	})
}
