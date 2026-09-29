package services

import (
	"context"
	"net/url"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type InterviewService struct {
	repo     ports.InterviewRepository
	issues   ports.IssueRepository
	projects ports.ProjectRepository
	domains  ports.DomainRepository
	guard    ports.Guard
	clock    ports.Clock
	ids      ports.IDGenerator
	log      ports.Logger
}

func NewInterviewService(repo ports.InterviewRepository, issues ports.IssueRepository, projects ports.ProjectRepository, domains ports.DomainRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *InterviewService {
	return &InterviewService{repo: repo, issues: issues, projects: projects, domains: domains, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.InterviewUseCase = (*InterviewService)(nil)

func (s *InterviewService) ticket(ctx context.Context, actor ports.Actor, id domain.IssueID, write bool) (domain.Issue, error) {
	ticket, err := s.issues.GetByID(ctx, id)
	if err != nil {
		return domain.Issue{}, err
	}
	if write {
		_, err = s.guard.EnsureWrite(ctx, actor, ticket.ProjectID)
	} else {
		_, err = s.guard.EnsureRead(ctx, actor, ticket.ProjectID)
	}
	return ticket, err
}

func (s *InterviewService) active(ctx context.Context, ticket domain.IssueID) (domain.Interview, bool, error) {
	list, err := s.repo.ListByIssue(ctx, ticket)
	if err != nil {
		return domain.Interview{}, false, err
	}
	for i := len(list) - 1; i >= 0; i-- {
		if !list[i].IsFinished() {
			return list[i], true, nil
		}
	}
	return domain.Interview{}, false, nil
}

func (s *InterviewService) activeFor(ctx context.Context, actor ports.Actor, id domain.IssueID, write bool) (domain.Interview, domain.Issue, error) {
	ticket, err := s.ticket(ctx, actor, id, write)
	if err != nil {
		return domain.Interview{}, domain.Issue{}, err
	}
	interview, ok, err := s.active(ctx, ticket.ID)
	if err != nil {
		return domain.Interview{}, domain.Issue{}, err
	}
	if !ok {
		return domain.Interview{}, domain.Issue{}, domain.ErrNotFound
	}
	return interview, ticket, nil
}

func (s *InterviewService) StartInterview(ctx context.Context, actor ports.Actor, id domain.IssueID) (domain.Interview, bool, error) {
	ticket, err := s.ticket(ctx, actor, id, true)
	if err != nil {
		return domain.Interview{}, false, err
	}
	if existing, ok, err := s.active(ctx, ticket.ID); err != nil || ok {
		return existing, false, err
	}
	if err := rules.CheckInterviewable(ticket); err != nil {
		return domain.Interview{}, false, err
	}

	now := s.clock.Now()
	created, err := s.repo.Create(ctx, domain.Interview{
		ID:          domain.InterviewID(s.ids.NewID()),
		ProjectID:   ticket.ProjectID,
		IssueID:     ticket.ID,
		Topic:       ticket.Title,
		State:       domain.InterviewState{Terms: []domain.Term{}, Questions: []domain.Question{}},
		AgentStatus: domain.AgentWaiting,
		AgentSince:  now,
		CreatedBy:   actor.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	return created, err == nil, err
}

func (s *InterviewService) CurrentInterview(ctx context.Context, actor ports.Actor, id domain.IssueID) (domain.Interview, error) {
	interview, _, err := s.activeFor(ctx, actor, id, false)
	return interview, err
}

func (s *InterviewService) ListInterviews(ctx context.Context, actor ports.Actor, id domain.IssueID) ([]domain.Interview, error) {
	ticket, err := s.ticket(ctx, actor, id, false)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByIssue(ctx, ticket.ID)
}

func (s *InterviewService) PatchInterview(ctx context.Context, actor ports.Actor, id domain.IssueID, patch []byte) (ports.InterviewPatchSummary, error) {
	interview, _, err := s.activeFor(ctx, actor, id, true)
	if err != nil {
		return ports.InterviewPatchSummary{}, err
	}
	now := s.clock.Now()
	next, handled, err := rules.ApplyInterviewPatch(interview.State, patch, now)
	if err != nil {
		return ports.InterviewPatchSummary{}, err
	}
	if handled != nil {
		sent, err := s.repo.EventsAfter(ctx, interview.ID, interview.Handled)
		if err != nil {
			return ports.InterviewPatchSummary{}, err
		}
		last := interview.Handled
		if len(sent) > 0 {
			last = sent[len(sent)-1].Seq
		}
		if err := rules.CheckHandled(interview.Handled, *handled, last); err != nil {
			return ports.InterviewPatchSummary{}, err
		}
		interview.Handled = *handled
		interview.AgentStatus = domain.AgentWaiting
		interview.AgentSince = now
	}

	summary := rules.SummarisePatch(interview.State, next)
	interview.State = next
	interview.UpdatedAt = now
	saved, err := s.repo.Update(ctx, interview)
	if err != nil {
		return ports.InterviewPatchSummary{}, err
	}
	return ports.InterviewPatchSummary{
		Interview: saved, Round: summary.Round, Added: summary.Added, Answered: summary.Answered, Handled: saved.Handled,
	}, nil
}

func (s *InterviewService) PagePath(ctx context.Context, actor ports.Actor, id domain.IssueID) (string, error) {
	ticket, err := s.ticket(ctx, actor, id, false)
	if err != nil {
		return "", err
	}
	project, err := s.projects.GetByID(ctx, ticket.ProjectID)
	if err != nil {
		return "", err
	}
	if project.DomainID == "" {
		return "", nil
	}
	dom, err := s.domains.GetByID(ctx, project.DomainID)
	if err != nil {
		return "", err
	}
	parts := []string{dom.ClientSlug, dom.Slug, project.Slug, "tickets", ticket.Slug, "interview"}
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return "/" + strings.Join(parts, "/"), nil
}

func (s *InterviewService) PendingSends(ctx context.Context, actor ports.Actor, id domain.IssueID) ([]domain.InterviewEvent, error) {
	interview, _, err := s.activeFor(ctx, actor, id, true)
	if err != nil {
		return nil, err
	}
	pending, err := s.repo.EventsAfter(ctx, interview.ID, interview.Handled)
	if err != nil {
		return nil, err
	}
	if len(pending) > 0 && interview.AgentStatus != domain.AgentWorking {
		now := s.clock.Now()
		interview.AgentStatus = domain.AgentWorking
		interview.AgentSince = now
		interview.UpdatedAt = now
		if _, err := s.repo.Update(ctx, interview); err != nil {
			return nil, err
		}
	}
	return pending, nil
}

func (s *InterviewService) SendToInterview(ctx context.Context, actor ports.Actor, id domain.IssueID, actions []domain.SendAction) (domain.InterviewEvent, error) {
	interview, _, err := s.activeFor(ctx, actor, id, true)
	if err != nil {
		return domain.InterviewEvent{}, err
	}
	if err := rules.ValidateSend(interview.State, actions); err != nil {
		return domain.InterviewEvent{}, err
	}
	return s.repo.AppendEvent(ctx, interview, actions, s.clock.Now())
}

func (s *InterviewService) FinishInterview(ctx context.Context, actor ports.Actor, id domain.IssueID, answer, doc string) (domain.Interview, error) {
	interview, ticket, err := s.activeFor(ctx, actor, id, true)
	if err != nil {
		return domain.Interview{}, err
	}
	if err := rules.CheckFinishable(interview.State); err != nil {
		return domain.Interview{}, err
	}

	now := s.clock.Now()
	resolution := domain.Entry{
		ID:        domain.EntryID(s.ids.NewID()),
		Kind:      domain.EntryResolution,
		ProjectID: ticket.ProjectID,
		IssueID:   ticket.ID,
		Body:      doc,
		CreatedBy: actor.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rules.ValidateEntry(resolution); err != nil {
		return domain.Interview{}, err
	}

	if err := rules.CanTransitionIssue(ticket.Status, domain.IssueDone); err != nil {
		return domain.Interview{}, err
	}
	ticket = rules.ApplyIssueStatus(ticket, domain.IssueDone)
	ticket.Resolution = strings.TrimSpace(answer)
	ticket.ResolutionEntry = resolution.ID
	ticket.UpdatedAt = now
	if err := rules.ValidateIssue(ticket); err != nil {
		return domain.Interview{}, err
	}
	if err := rules.CheckResolvedIssue(ticket, ticket.Status); err != nil {
		return domain.Interview{}, err
	}

	sent, err := s.repo.EventsAfter(ctx, interview.ID, interview.Handled)
	if err != nil {
		return domain.Interview{}, err
	}
	if len(sent) > 0 {
		interview.Handled = sent[len(sent)-1].Seq
	}
	interview.FinishedAt = &now
	interview.AgentStatus = domain.AgentWaiting
	interview.AgentSince = now
	interview.UpdatedAt = now
	if err := s.repo.Finish(ctx, interview, ticket, resolution); err != nil {
		return domain.Interview{}, err
	}
	return interview, nil
}
