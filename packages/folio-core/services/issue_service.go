package services

import (
	"context"
	"sort"
	"strings"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type IssueService struct {
	repo    ports.IssueRepository
	plans   ports.PlanRepository
	journal ports.JournalRepository
	docs    ports.DocRepository
	cycles  ports.CycleRepository
	guard   ports.Guard
	clock   ports.Clock
	ids     ports.IDGenerator
	log     ports.Logger
}

func NewIssueService(repo ports.IssueRepository, plans ports.PlanRepository, journal ports.JournalRepository, docs ports.DocRepository, cycles ports.CycleRepository, guard ports.Guard, clock ports.Clock, ids ports.IDGenerator, log ports.Logger) *IssueService {
	return &IssueService{repo: repo, plans: plans, journal: journal, docs: docs, cycles: cycles, guard: guard, clock: clock, ids: ids, log: log}
}

var _ ports.IssueUseCase = (*IssueService)(nil)

func (s *IssueService) ListIssues(ctx context.Context, actor ports.Actor, project domain.ProjectID, f domain.IssueFilter) ([]domain.Issue, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return nil, err
	}
	f.Limit = clampLimit(f.Limit)
	issues, err := s.repo.List(ctx, project, f)
	if err != nil {
		return nil, err
	}
	if err := s.decorate(ctx, project, issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// decorate resolves the link-derived fields against the whole project, since a
// blocker may sit outside the filtered page.
func (s *IssueService) decorate(ctx context.Context, project domain.ProjectID, issues []domain.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	links, err := s.repo.LinksOfProject(ctx, project)
	if err != nil {
		return err
	}
	all, err := s.repo.List(ctx, project, domain.IssueFilter{Limit: MaxPageSize})
	if err != nil {
		return err
	}
	rules.ApplyLinks(all, links)

	derived := make(map[domain.IssueID]domain.Issue, len(all))
	for _, i := range all {
		derived[i.ID] = i
	}
	for idx := range issues {
		if d, ok := derived[issues[idx].ID]; ok {
			issues[idx].ParentID = d.ParentID
			issues[idx].DependsOn = d.DependsOn
			issues[idx].RelatedTo = d.RelatedTo
			issues[idx].Blocked = d.Blocked
		}
		if issues[idx].Kind == domain.IssueTicket {
			issues[idx].Progress = s.progress(issues[idx].ID, all, links)
		}
	}
	return nil
}

func (s *IssueService) progress(id domain.IssueID, all []domain.Issue, links []domain.IssueLink) domain.Progress {
	children := make([]domain.Issue, 0)
	byID := make(map[domain.IssueID]domain.Issue, len(all))
	for _, i := range all {
		byID[i.ID] = i
	}
	for _, l := range links {
		if l.Kind == domain.LinkParent && l.To == id {
			if child, ok := byID[l.From]; ok {
				children = append(children, child)
			}
		}
	}
	return rules.ProgressOfIssues(children)
}

func (s *IssueService) GetIssue(ctx context.Context, actor ports.Actor, id domain.IssueID) (domain.Issue, error) {
	issue, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Issue{}, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, issue.ProjectID); err != nil {
		return domain.Issue{}, err
	}
	return s.hydrate(ctx, issue)
}

func (s *IssueService) GetIssueBySlug(ctx context.Context, actor ports.Actor, project domain.ProjectID, slug string) (domain.Issue, error) {
	if _, err := s.guard.EnsureRead(ctx, actor, project); err != nil {
		return domain.Issue{}, err
	}
	issue, err := s.repo.GetBySlug(ctx, project, slug)
	if err != nil {
		return domain.Issue{}, err
	}
	return s.hydrate(ctx, issue)
}

func (s *IssueService) hydrate(ctx context.Context, issue domain.Issue) (domain.Issue, error) {
	one := []domain.Issue{issue}
	if err := s.decorate(ctx, issue.ProjectID, one); err != nil {
		return domain.Issue{}, err
	}
	issue = one[0]

	if issue.Kind == domain.IssueTicket {
		cycles, err := s.cycles.ListByTicket(ctx, domain.TicketID(issue.ID))
		if err != nil {
			return domain.Issue{}, err
		}
		if current, ok := rules.CurrentCycle(cycles); ok {
			issue.Cycle, issue.Phase = current.Ordinal, current.Phase
		}
	}
	return issue, nil
}

func (s *IssueService) CreateIssue(ctx context.Context, actor ports.Actor, in ports.CreateIssueInput) (domain.Issue, error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, in.ProjectID); err != nil {
		return domain.Issue{}, err
	}
	return s.create(ctx, actor, in, 0)
}

func (s *IssueService) create(ctx context.Context, actor ports.Actor, in ports.CreateIssueInput, position int) (domain.Issue, error) {
	if err := s.checkPlan(ctx, in.ProjectID, in.PlanID); err != nil {
		return domain.Issue{}, err
	}
	due, err := parseDue(in.DueDate)
	if err != nil {
		return domain.Issue{}, err
	}

	kind := in.Kind
	if kind == "" {
		kind = domain.IssueTodo
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = rules.Slugify(in.Title)
	}
	if position == 0 && kind == domain.IssueTodo {
		existing, err := s.repo.List(ctx, in.ProjectID, domain.IssueFilter{PlanID: in.PlanID, Limit: MaxPageSize})
		if err != nil {
			return domain.Issue{}, err
		}
		position = rules.NextIssuePosition(existing)
	}

	now := s.clock.Now()
	issue := domain.Issue{
		ID:          domain.IssueID(s.ids.NewID()),
		Kind:        kind,
		ProjectID:   in.ProjectID,
		PlanID:      in.PlanID,
		Slug:        slug,
		Title:       strings.TrimSpace(in.Title),
		Body:        in.Body,
		Status:      defaultIssueStatus(in.Status),
		Priority:    defaultPriority(in.Priority),
		Size:        in.Size,
		Assignee:    in.Assignee,
		Tags:        in.Tags,
		Position:    position,
		DueDate:     due,
		Wayfinder:   in.Wayfinder,
		ExternalRef: in.ExternalRef,
		CreatedBy:   actor.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := rules.ValidateIssue(issue); err != nil {
		return domain.Issue{}, err
	}
	if err := s.slugFree(ctx, issue.ProjectID, issue.Slug, ""); err != nil {
		return domain.Issue{}, err
	}

	created, err := s.repo.Create(ctx, issue)
	if err != nil {
		return domain.Issue{}, err
	}

	if in.ParentID != "" {
		if err := s.link(ctx, created.ID, in.ParentID, domain.LinkParent, created.ProjectID); err != nil {
			return domain.Issue{}, err
		}
	}
	for _, dep := range in.DependsOn {
		if err := s.link(ctx, dep, created.ID, domain.LinkBlocks, created.ProjectID); err != nil {
			return domain.Issue{}, err
		}
	}
	return s.hydrate(ctx, created)
}

// link refuses an edge that leaves the project or closes a loop: a cross-project
// edge would expose work the caller cannot read.
func (s *IssueService) link(ctx context.Context, from, to domain.IssueID, kind domain.LinkKind, project domain.ProjectID) error {
	if err := s.inScope(ctx, from, project); err != nil {
		return err
	}
	if err := s.inScope(ctx, to, project); err != nil {
		return err
	}
	links, err := s.repo.LinksOfProject(ctx, project)
	if err != nil {
		return err
	}
	if err := rules.CheckNoLinkCycle(from, to, kind, links); err != nil {
		return err
	}
	return s.repo.Link(ctx, from, to, kind)
}

func (s *IssueService) inScope(ctx context.Context, id domain.IssueID, project domain.ProjectID) error {
	if id == "" {
		return nil
	}
	issue, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if notFound(err) {
			return domain.Invalid("issue", "does not exist")
		}
		return err
	}
	if issue.ProjectID != project {
		return domain.Invalid("issue", "belongs to a different project")
	}
	return nil
}

func (s *IssueService) checkPlan(ctx context.Context, project domain.ProjectID, plan domain.PlanID) error {
	if plan == "" {
		return nil
	}
	found, err := s.plans.GetByID(ctx, plan)
	if err != nil {
		if notFound(err) {
			return domain.Invalid("plan", "does not exist")
		}
		return err
	}
	if found.ProjectID != project {
		return domain.Invalid("plan", "belongs to a different project")
	}
	return nil
}

func (s *IssueService) slugFree(ctx context.Context, project domain.ProjectID, slug string, except domain.IssueID) error {
	existing, err := s.repo.GetBySlug(ctx, project, slug)
	if err != nil {
		if notFound(err) {
			return nil
		}
		return err
	}
	if existing.ID == except {
		return nil
	}
	return domain.ErrConflict
}

func (s *IssueService) CreateIssues(ctx context.Context, actor ports.Actor, project domain.ProjectID, in []ports.CreateIssueInput) (ports.BatchResult[domain.Issue], error) {
	if _, err := s.guard.EnsureWrite(ctx, actor, project); err != nil {
		return ports.BatchResult[domain.Issue]{}, err
	}
	existing, err := s.repo.List(ctx, project, domain.IssueFilter{Limit: MaxPageSize})
	if err != nil {
		return ports.BatchResult[domain.Issue]{}, err
	}
	next := rules.NextIssuePosition(existing)

	out := ports.BatchResult[domain.Issue]{Created: make([]domain.Issue, 0, len(in))}
	for i, item := range in {
		item.ProjectID = project
		created, err := s.create(ctx, actor, item, next)
		if err != nil {
			out.Errors = append(out.Errors, ports.BatchError{Index: i, Title: item.Title, Reason: err.Error()})
			continue
		}
		out.Created = append(out.Created, created)
		next++
	}
	return out, nil
}

func (s *IssueService) UpdateIssue(ctx context.Context, actor ports.Actor, id domain.IssueID, in ports.UpdateIssueInput) (domain.Issue, error) {
	issue, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Issue{}, err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, issue.ProjectID); err != nil {
		return domain.Issue{}, err
	}

	if in.Status != nil {
		if err := rules.CanTransitionIssue(issue.Status, *in.Status); err != nil {
			return domain.Issue{}, err
		}
		if in.Status.IsTerminal() && issue.Kind == domain.IssueTicket {
			cycles, err := s.cycles.ListByTicket(ctx, domain.TicketID(issue.ID))
			if err != nil {
				return domain.Issue{}, err
			}
			if err := rules.CheckClosableIssue(*in.Status, cycles); err != nil {
				return domain.Issue{}, err
			}
		}
		issue.Status = *in.Status
	}
	if in.Kind != nil {
		issue.Kind = *in.Kind
	}
	if in.PlanID != nil {
		if err := s.checkPlan(ctx, issue.ProjectID, *in.PlanID); err != nil {
			return domain.Issue{}, err
		}
		issue.PlanID = *in.PlanID
	}
	if in.Slug != nil {
		issue.Slug = strings.TrimSpace(*in.Slug)
	}
	if in.Title != nil {
		issue.Title = strings.TrimSpace(*in.Title)
	}
	if in.Body != nil {
		issue.Body = *in.Body
	}
	if in.Priority != nil {
		issue.Priority = *in.Priority
	}
	if in.Size != nil {
		issue.Size = *in.Size
	}
	if in.Assignee != nil {
		issue.Assignee = *in.Assignee
	}
	if in.Tags != nil {
		issue.Tags = *in.Tags
	}
	if in.Position != nil {
		issue.Position = *in.Position
	}
	if in.Wayfinder != nil {
		issue.Wayfinder = *in.Wayfinder
	}
	if in.ExternalRef != nil {
		issue.ExternalRef = *in.ExternalRef
	}
	if in.DueDate != nil {
		due, err := parseDue(in.DueDate)
		if err != nil {
			return domain.Issue{}, err
		}
		issue.DueDate = due
	}

	if err := rules.ValidateIssue(issue); err != nil {
		return domain.Issue{}, err
	}
	if in.Slug != nil {
		if err := s.slugFree(ctx, issue.ProjectID, issue.Slug, issue.ID); err != nil {
			return domain.Issue{}, err
		}
	}

	issue.UpdatedAt = s.clock.Now()
	saved, err := s.repo.Update(ctx, issue)
	if err != nil {
		return domain.Issue{}, err
	}

	if in.ParentID != nil {
		if err := s.setParent(ctx, saved, *in.ParentID); err != nil {
			return domain.Issue{}, err
		}
	}
	if in.DependsOn != nil {
		for _, dep := range *in.DependsOn {
			if err := s.inScope(ctx, dep, saved.ProjectID); err != nil {
				return domain.Issue{}, err
			}
		}
		if err := s.setBlockers(ctx, saved, *in.DependsOn); err != nil {
			return domain.Issue{}, err
		}
	}
	return s.hydrate(ctx, saved)
}

func (s *IssueService) setParent(ctx context.Context, issue domain.Issue, parent domain.IssueID) error {
	links, err := s.repo.Links(ctx, issue.ID)
	if err != nil {
		return err
	}
	for _, l := range links {
		if l.Kind == domain.LinkParent && l.From == issue.ID {
			if l.To == parent {
				return nil
			}
			if err := s.repo.Unlink(ctx, l.From, l.To, domain.LinkParent); err != nil {
				return err
			}
		}
	}
	if parent == "" {
		return nil
	}
	return s.link(ctx, issue.ID, parent, domain.LinkParent, issue.ProjectID)
}

func (s *IssueService) setBlockers(ctx context.Context, issue domain.Issue, deps []domain.IssueID) error {
	links, err := s.repo.LinksOfProject(ctx, issue.ProjectID)
	if err != nil {
		return err
	}
	for _, dep := range deps {
		if err := rules.CheckNoLinkCycle(dep, issue.ID, domain.LinkBlocks, links); err != nil {
			return err
		}
	}

	current, err := s.repo.Links(ctx, issue.ID)
	if err != nil {
		return err
	}
	want := make(map[domain.IssueID]bool, len(deps))
	for _, d := range deps {
		want[d] = true
	}
	for _, l := range current {
		if l.Kind != domain.LinkBlocks || l.To != issue.ID {
			continue
		}
		if want[l.From] {
			delete(want, l.From)
			continue
		}
		if err := s.repo.Unlink(ctx, l.From, l.To, domain.LinkBlocks); err != nil {
			return err
		}
	}
	for dep := range want {
		if err := s.repo.Link(ctx, dep, issue.ID, domain.LinkBlocks); err != nil {
			return err
		}
	}
	return nil
}

func (s *IssueService) SetIssueStatus(ctx context.Context, actor ports.Actor, id domain.IssueID, status domain.IssueStatus) (domain.Issue, error) {
	return s.UpdateIssue(ctx, actor, id, ports.UpdateIssueInput{Status: &status})
}

func (s *IssueService) LinkIssues(ctx context.Context, actor ports.Actor, from, to domain.IssueID, kind domain.LinkKind) error {
	issue, err := s.repo.GetByID(ctx, from)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, issue.ProjectID); err != nil {
		return err
	}
	return s.link(ctx, from, to, kind, issue.ProjectID)
}

func (s *IssueService) UnlinkIssues(ctx context.Context, actor ports.Actor, from, to domain.IssueID, kind domain.LinkKind) error {
	issue, err := s.repo.GetByID(ctx, from)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, issue.ProjectID); err != nil {
		return err
	}
	return s.repo.Unlink(ctx, from, to, kind)
}

// DeleteIssue detaches its children rather than deleting them: the issue is a
// framing of the work, and dropping it should not destroy the work itself.
func (s *IssueService) DeleteIssue(ctx context.Context, actor ports.Actor, id domain.IssueID) error {
	issue, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.guard.EnsureWrite(ctx, actor, issue.ProjectID); err != nil {
		return err
	}
	if err := s.detachChildren(ctx, issue); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *IssueService) detachChildren(ctx context.Context, issue domain.Issue) error {
	now := s.clock.Now()

	plans, err := s.plans.ListByTicket(ctx, domain.TicketID(issue.ID))
	if err != nil {
		return err
	}
	for _, p := range plans {
		p.TicketID = ""
		p.UpdatedAt = now
		if _, err := s.plans.Update(ctx, p); err != nil {
			return err
		}
	}

	entries, err := s.journal.List(ctx, issue.ProjectID, domain.JournalFilter{TicketID: domain.TicketID(issue.ID), Limit: MaxPageSize})
	if err != nil {
		return err
	}
	for _, e := range entries {
		e.TicketID = ""
		e.UpdatedAt = now
		if _, err := s.journal.Update(ctx, e); err != nil {
			return err
		}
	}

	docs, err := s.docs.List(ctx, issue.ProjectID, domain.DocFilter{TicketID: domain.TicketID(issue.ID), Limit: MaxPageSize})
	if err != nil {
		return err
	}
	for _, d := range docs {
		d.TicketID = ""
		d.UpdatedAt = now
		if _, err := s.docs.Update(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

func (s *IssueService) Frontier(ctx context.Context, actor ports.Actor, mapID domain.IssueID) ([]domain.Issue, error) {
	parent, err := s.repo.GetByID(ctx, mapID)
	if err != nil {
		return nil, err
	}
	if _, err := s.guard.EnsureRead(ctx, actor, parent.ProjectID); err != nil {
		return nil, err
	}
	children, err := s.repo.ListByParent(ctx, mapID)
	if err != nil {
		return nil, err
	}
	if err := s.decorate(ctx, parent.ProjectID, children); err != nil {
		return nil, err
	}

	out := make([]domain.Issue, 0, len(children))
	for _, child := range children {
		if child.Status.IsTerminal() || child.Assignee != "" || child.Blocked {
			continue
		}
		out = append(out, child)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (s *IssueService) GetIssueBrief(ctx context.Context, actor ports.Actor, id domain.IssueID, in ports.BriefOptions) (domain.IssueBrief, error) {
	issue, err := s.GetIssue(ctx, actor, id)
	if err != nil {
		return domain.IssueBrief{}, err
	}
	return s.brief(ctx, issue, in)
}

func (s *IssueService) GetIssueBriefBySlug(ctx context.Context, actor ports.Actor, project domain.ProjectID, slug string, in ports.BriefOptions) (domain.IssueBrief, error) {
	issue, err := s.GetIssueBySlug(ctx, actor, project, slug)
	if err != nil {
		return domain.IssueBrief{}, err
	}
	return s.brief(ctx, issue, in)
}

func (s *IssueService) brief(ctx context.Context, issue domain.Issue, in ports.BriefOptions) (domain.IssueBrief, error) {
	recent := in.RecentJournal
	if recent <= 0 {
		recent = DefaultRecentJournal
	}

	children, err := s.repo.ListByParent(ctx, issue.ID)
	if err != nil {
		return domain.IssueBrief{}, err
	}
	if err := s.decorate(ctx, issue.ProjectID, children); err != nil {
		return domain.IssueBrief{}, err
	}
	sortOpenFirstIssues(children)

	plans, err := s.plans.ListByTicket(ctx, domain.TicketID(issue.ID))
	if err != nil {
		return domain.IssueBrief{}, err
	}

	journal, err := s.journal.List(ctx, issue.ProjectID, domain.JournalFilter{TicketID: domain.TicketID(issue.ID), Limit: recent})
	if err != nil {
		return domain.IssueBrief{}, err
	}

	docs, err := s.docs.List(ctx, issue.ProjectID, domain.DocFilter{TicketID: domain.TicketID(issue.ID), Limit: MaxPageSize})
	if err != nil {
		return domain.IssueBrief{}, err
	}

	cycles, err := s.cycles.ListByTicket(ctx, domain.TicketID(issue.ID))
	if err != nil {
		return domain.IssueBrief{}, err
	}
	if cycles == nil {
		cycles = []domain.Cycle{}
	}

	return domain.IssueBrief{
		Issue:    issue,
		Children: children,
		Plans:    plans,
		Journal:  journal,
		Docs:     docs,
		Cycles:   cycles,
	}, nil
}

func sortOpenFirstIssues(issues []domain.Issue) {
	sort.SliceStable(issues, func(i, j int) bool {
		return !issues[i].Status.IsTerminal() && issues[j].Status.IsTerminal()
	})
}

func defaultIssueStatus(s domain.IssueStatus) domain.IssueStatus {
	if s == "" {
		return domain.IssueOpen
	}
	return s
}
