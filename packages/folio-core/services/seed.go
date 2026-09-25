package services

import (
	"context"
	"errors"
	"fmt"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// SeedUseCases is the slice of the hexagon a seed needs. Taking the use case
// ports rather than the repositories means seeded data passes the same
// validation and permission checks as anything a client writes, so a seed
// that succeeds proves the API works.
type SeedUseCases struct {
	Projects ports.ProjectUseCase
	Plans    ports.PlanUseCase
	Issues   ports.IssueUseCase
	Entries  ports.EntryUseCase
	Cycles   ports.CycleUseCase
}

// SeedReport counts what a seed run wrote.
type SeedReport struct {
	Projects int
	Tickets  int
	Plans    int
	Todos    int
	Journal  int
	Docs     int
	Cycles   int
	WorkLogs int
}

func (r SeedReport) String() string {
	return fmt.Sprintf("%d projects, %d tickets, %d plans, %d todos, %d journal, %d docs, %d cycles, %d work logs",
		r.Projects, r.Tickets, r.Plans, r.Todos, r.Journal, r.Docs, r.Cycles, r.WorkLogs)
}

// Seed writes the demo dataset as the given actor, who becomes the owner of
// every project it creates. It is not idempotent: a project whose slug is
// taken is reused rather than recreated, so a second run adds this dataset's
// contents to it and duplicates whatever it wrote before.
func Seed(ctx context.Context, uc SeedUseCases, actor ports.Actor) (SeedReport, error) {
	var report SeedReport

	for _, spec := range seedProjects() {
		project, err := uc.Projects.CreateProject(ctx, actor, spec.project)
		if err != nil {
			if !errors.Is(err, domain.ErrConflict) {
				return report, fmt.Errorf("project %q: %w", spec.project.Slug, err)
			}
			if project, err = uc.Projects.GetProject(ctx, actor, spec.project.Slug); err != nil {
				return report, fmt.Errorf("project %q: %w", spec.project.Slug, err)
			}
		} else {
			report.Projects++
		}

		// Tickets come first: the plans below hang off the first one, so the
		// demo shows work organised under a ticket rather than only loose
		// under the project.
		issueIDs := make([]domain.IssueID, 0, len(spec.tickets))
		for _, ticket := range spec.tickets {
			ticket.ProjectID = project.ID
			created, err := uc.Issues.CreateIssue(ctx, actor, ticket)
			if err != nil {
				if errors.Is(err, domain.ErrConflict) {
					continue
				}
				return report, fmt.Errorf("ticket %q: %w", ticket.Title, err)
			}
			report.Tickets++
			issueIDs = append(issueIDs, created.ID)
		}

		// Plans are created with their todos in one call, the same way an
		// agent drafting a plan would.
		planIDs := make([]domain.PlanID, 0, len(spec.plans))
		for i, plan := range spec.plans {
			plan.ProjectID = project.ID
			// Only the first plan is filed under the ticket, so the dataset
			// covers both a ticket with work under it and work that has none.
			if i == 0 && len(issueIDs) > 0 {
				plan.IssueID = issueIDs[0]
			}
			created, err := uc.Plans.CreatePlan(ctx, actor, plan)
			if err != nil {
				return report, fmt.Errorf("plan %q: %w", plan.Title, err)
			}
			report.Plans++
			report.Todos += len(plan.Todos)
			planIDs = append(planIDs, created.ID)
		}

		// Advance some todos so plans show real progress rather than 0%.
		for i, planID := range planIDs {
			if i >= len(spec.done) {
				break
			}
			todos, err := uc.Issues.ListIssues(ctx, actor, project.ID, domain.IssueFilter{PlanID: planID})
			if err != nil {
				return report, err
			}
			for j, status := range spec.done[i] {
				if j >= len(todos) {
					break
				}
				if _, err := uc.Issues.SetIssueStatus(ctx, actor, todos[j].ID, status); err != nil {
					return report, fmt.Errorf("todo %q: %w", todos[j].Title, err)
				}
			}
		}

		for _, entry := range spec.journal {
			entry.ProjectID = project.ID
			entry.Kind = domain.EntryJournal
			if len(planIDs) > 0 && entry.PlanID == "" {
				entry.PlanID = planIDs[0]
			}
			if _, err := uc.Entries.WriteEntry(ctx, actor, entry); err != nil {
				return report, fmt.Errorf("log %q: %w", entry.Title, err)
			}
			report.Journal++
		}

		for _, doc := range spec.docs {
			doc.ProjectID = project.ID
			doc.Kind = domain.EntryDoc
			if _, err := uc.Entries.WriteEntry(ctx, actor, doc); err != nil {
				if errors.Is(err, domain.ErrConflict) {
					continue
				}
				return report, fmt.Errorf("doc %q: %w", doc.Title, err)
			}
			report.Docs++
		}

		if spec.wayfinder != nil {
			if err := seedWayfinder(ctx, uc, actor, project.ID, *spec.wayfinder, &report); err != nil {
				return report, err
			}
		}
	}

	return report, nil
}

// seedWayfinder writes a map ticket and the graph under it. Keys resolve to
// ids as it goes, so a node can name a parent or blocker declared above it;
// the spec is ordered so that always holds.
func seedWayfinder(
	ctx context.Context,
	uc SeedUseCases,
	actor ports.Actor,
	project domain.ProjectID,
	spec wayfinderSpec,
	report *SeedReport,
) error {
	ids := make(map[string]domain.IssueID, len(spec.nodes)+1)

	for _, node := range append([]wayfinderNode{spec.root}, spec.nodes...) {
		in := node.issue
		in.ProjectID = project
		in.Kind = domain.IssueTicket

		if node.parent != "" {
			parent, ok := ids[node.parent]
			if !ok {
				return fmt.Errorf("wayfinder %q: parent %q is not written yet", node.key, node.parent)
			}
			in.ParentID = parent
		}

		for _, key := range node.dependsOn {
			blocker, ok := ids[key]
			if !ok {
				return fmt.Errorf("wayfinder %q: blocker %q is not written yet", node.key, key)
			}
			in.DependsOn = append(in.DependsOn, blocker)
		}

		created, err := uc.Issues.CreateIssue(ctx, actor, in)
		if err != nil {
			if !errors.Is(err, domain.ErrConflict) {
				return fmt.Errorf("wayfinder %q: %w", node.key, err)
			}
			// Already seeded. Its id still has to reach the nodes below that
			// name it as a parent or blocker, and its cycles and relations are
			// already written, so only the todos are worth revisiting.
			existing, err := uc.Issues.GetIssueBySlug(ctx, actor, project, in.Slug)
			if err != nil {
				return fmt.Errorf("wayfinder %q: %w", node.key, err)
			}
			ids[node.key] = existing.ID
			if err := seedTodos(ctx, uc, actor, project, existing.ID, node, report); err != nil {
				return err
			}
			continue
		}
		report.Tickets++
		ids[node.key] = created.ID

		for _, key := range node.relatedTo {
			other, ok := ids[key]
			if !ok {
				return fmt.Errorf("wayfinder %q: relation %q is not written yet", node.key, key)
			}
			if err := uc.Issues.LinkIssues(ctx, actor, created.ID, other, domain.LinkRelates); err != nil {
				return fmt.Errorf("wayfinder %q: relate to %q: %w", node.key, key, err)
			}
		}

		if err := seedTodos(ctx, uc, actor, project, created.ID, node, report); err != nil {
			return err
		}

		if err := seedLogs(ctx, uc, actor, project, created.ID, node, report); err != nil {
			return err
		}

		if err := seedCycles(ctx, uc, actor, project, created.ID, node, report); err != nil {
			return err
		}
	}

	return nil
}

// seedTodos files the node's steps directly under its ticket. A todo carries
// no slug of its own, so a re-run cannot tell one apart by conflict the way a
// ticket can; it writes nothing when the ticket already has todos instead.
func seedTodos(
	ctx context.Context,
	uc SeedUseCases,
	actor ports.Actor,
	project domain.ProjectID,
	ticket domain.IssueID,
	node wayfinderNode,
	report *SeedReport,
) error {
	if len(node.todos) == 0 {
		return nil
	}

	existing, err := uc.Issues.ListIssues(ctx, actor, project, domain.IssueFilter{Kind: domain.IssueTodo, ParentID: ticket})
	if err != nil {
		return fmt.Errorf("wayfinder %q: list todos: %w", node.key, err)
	}
	if len(existing) > 0 {
		return nil
	}

	for _, todo := range node.todos {
		todo.ProjectID = project
		todo.Kind = domain.IssueTodo
		todo.ParentID = ticket
		if _, err := uc.Issues.CreateIssue(ctx, actor, todo); err != nil {
			return fmt.Errorf("wayfinder %q: todo %q: %w", node.key, todo.Title, err)
		}
		report.Todos++
	}

	return nil
}

var phaseOrder = []domain.Phase{domain.PhasePlan, domain.PhaseDo, domain.PhaseCheck, domain.PhaseAct}

// seedCycles replays the PDCA rounds through the API rather than writing the
// end state: phases only advance one step, and a round has to be resolved
// before the next can open, so the sequence is the only way to reach it.
func seedCycles(
	ctx context.Context,
	uc SeedUseCases,
	actor ports.Actor,
	project domain.ProjectID,
	ticket domain.IssueID,
	node wayfinderNode,
	report *SeedReport,
) error {
	for _, round := range node.cycles {
		cycle, err := uc.Cycles.OpenCycle(ctx, actor, ticket)
		if err != nil {
			return fmt.Errorf("wayfinder %q: open cycle: %w", node.key, err)
		}
		report.Cycles++

		for _, phase := range phaseOrder[1:] {
			if cycle.Phase == round.phase {
				break
			}
			if cycle, err = uc.Cycles.AdvancePhase(ctx, actor, cycle.ID, phase); err != nil {
				return fmt.Errorf("wayfinder %q: advance to %s: %w", node.key, phase, err)
			}
		}

		// Written before the round is resolved, since a work log is stamped
		// with the cycle that is open when it lands.
		for i, log := range round.logs {
			log.ProjectID = project
			log.Kind = domain.EntryLog
			log.IssueID = ticket
			log.CycleID = cycle.ID
			if log.Title == "" {
				log.Title = fmt.Sprintf("%s, note %d", node.issue.Title, i+1)
			}
			if _, err := uc.Entries.WriteEntry(ctx, actor, log); err != nil {
				return fmt.Errorf("wayfinder %q: work log: %w", node.key, err)
			}
			report.WorkLogs++
		}

		if round.resolution != "" {
			if _, err := uc.Cycles.ResolveCycle(ctx, actor, cycle.ID, round.resolution); err != nil {
				return fmt.Errorf("wayfinder %q: resolve cycle: %w", node.key, err)
			}
		}
	}

	return nil
}

func seedLogs(
	ctx context.Context,
	uc SeedUseCases,
	actor ports.Actor,
	project domain.ProjectID,
	ticket domain.IssueID,
	node wayfinderNode,
	report *SeedReport,
) error {
	for i, log := range node.logs {
		log.ProjectID = project
		log.Kind = domain.EntryLog
		log.IssueID = ticket
		if log.Title == "" {
			log.Title = fmt.Sprintf("%s, note %d", node.issue.Title, i+1)
		}
		if _, err := uc.Entries.WriteEntry(ctx, actor, log); err != nil {
			return fmt.Errorf("wayfinder %q: work log: %w", node.key, err)
		}
		report.WorkLogs++
	}
	return nil
}
