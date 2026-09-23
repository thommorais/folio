package services

import (
	"context"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type ShareService struct {
	repo    ports.ShareRepository
	issues  ports.IssueRepository
	plans   ports.PlanRepository
	issueUC ports.IssueUseCase
	planUC  ports.PlanUseCase
	guard   ports.Guard
	clock   ports.Clock
	ids     ports.IDGenerator
	tokens  ports.TokenGenerator
	log     ports.Logger
}

func NewShareService(repo ports.ShareRepository, issues ports.IssueRepository, plans ports.PlanRepository, issueUC ports.IssueUseCase, planUC ports.PlanUseCase, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, tokens ports.TokenGenerator, log ports.Logger) *ShareService {
	return &ShareService{repo: repo, issues: issues, plans: plans, issueUC: issueUC, planUC: planUC, guard: guard, clock: clock, ids: ids, tokens: tokens, log: log}
}

var _ ports.ShareUseCase = (*ShareService)(nil)

var holder = ports.Actor{Superuser: true}

func (s *ShareService) ShareIssue(ctx context.Context, actor ports.Actor, id domain.IssueID, label string) (domain.Share, error) {
	issue, err := s.issues.GetByID(ctx, id)
	if err != nil {
		return domain.Share{}, err
	}
	return s.create(ctx, actor, domain.Share{ProjectID: issue.ProjectID, IssueID: issue.ID, Label: label})
}

func (s *ShareService) SharePlan(ctx context.Context, actor ports.Actor, id domain.PlanID, label string) (domain.Share, error) {
	plan, err := s.plans.GetByID(ctx, id)
	if err != nil {
		return domain.Share{}, err
	}
	return s.create(ctx, actor, domain.Share{ProjectID: plan.ProjectID, PlanID: plan.ID, Label: label})
}

func (s *ShareService) create(ctx context.Context, actor ports.Actor, share domain.Share) (domain.Share, error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, share.ProjectID); err != nil {
		return domain.Share{}, err
	}
	share.ID = domain.ShareID(s.ids.NewID())
	share.Label = strings.TrimSpace(share.Label)
	share.Token = s.tokens.NewToken()
	share.CreatedBy = actor.UserID
	share.CreatedAt = s.clock.Now()
	if err := rules.ValidateShare(share); err != nil {
		return domain.Share{}, err
	}
	return s.repo.Create(ctx, share)
}

func (s *ShareService) ListShares(ctx context.Context, actor ports.Actor, project domain.ProjectID) ([]domain.Share, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, project)
}

func (s *ShareService) RevokeShare(ctx context.Context, actor ports.Actor, id domain.ShareID) error {
	share, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, share.ProjectID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *ShareService) OpenShare(ctx context.Context, token string) (domain.SharedItem, error) {
	if token == "" {
		return domain.SharedItem{}, domain.ErrNotFound
	}
	share, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return domain.SharedItem{}, err
	}

	item := domain.SharedItem{Share: share}
	if share.IssueID != "" {
		brief, err := s.issueUC.GetIssueBrief(ctx, holder, share.IssueID, ports.BriefOptions{})
		if err != nil {
			return domain.SharedItem{}, err
		}
		item.Brief = &brief
	} else {
		plan, err := s.planUC.GetPlan(ctx, holder, share.PlanID)
		if err != nil {
			return domain.SharedItem{}, err
		}
		todos, err := s.issueUC.ListIssues(ctx, holder, plan.ProjectID, domain.IssueFilter{PlanID: plan.ID, Limit: MaxPageSize})
		if err != nil {
			return domain.SharedItem{}, err
		}
		item.Plan = &plan
		item.Todos = todos
	}

	if err := s.repo.Touch(ctx, share.ID, s.clock.Now()); err != nil {
		s.log.Warn("share visit not recorded", map[string]any{"share": string(share.ID), "error": err.Error()})
	}
	return item, nil
}
