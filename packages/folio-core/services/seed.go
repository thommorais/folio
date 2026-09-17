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
	Journal  ports.JournalUseCase
	Docs     ports.DocUseCase
}

// SeedReport counts what a seed run wrote.
type SeedReport struct {
	Projects int
	Tickets  int
	Plans    int
	Todos    int
	Journal  int
	Docs     int
}

func (r SeedReport) String() string {
	return fmt.Sprintf("%d projects, %d tickets, %d plans, %d todos, %d journal, %d docs",
		r.Projects, r.Tickets, r.Plans, r.Todos, r.Journal, r.Docs)
}

// Seed writes the demo dataset as the given actor, who becomes the owner of
// every project it creates. It is not idempotent: a project whose slug is
// taken is skipped, and its contents are skipped with it.
func Seed(ctx context.Context, uc SeedUseCases, actor ports.Actor) (SeedReport, error) {
	var report SeedReport

	for _, spec := range seedProjects() {
		project, err := uc.Projects.CreateProject(ctx, actor, spec.project)
		if err != nil {
			if errors.Is(err, domain.ErrConflict) {
				continue
			}
			return report, fmt.Errorf("project %q: %w", spec.project.Slug, err)
		}
		report.Projects++

		// Tickets come first: the plans below hang off the first one, so the
		// demo shows work organised under a ticket rather than only loose
		// under the project.
		issueIDs := make([]domain.IssueID, 0, len(spec.tickets))
		for _, ticket := range spec.tickets {
			ticket.ProjectID = project.ID
			created, err := uc.Issues.CreateIssue(ctx, actor, ticket)
			if err != nil {
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
			if len(planIDs) > 0 && entry.PlanID == "" {
				entry.PlanID = planIDs[0]
			}
			if _, err := uc.Journal.WriteJournalEntry(ctx, actor, entry); err != nil {
				return report, fmt.Errorf("log %q: %w", entry.Title, err)
			}
			report.Journal++
		}

		for _, doc := range spec.docs {
			doc.ProjectID = project.ID
			if _, err := uc.Docs.CreateDoc(ctx, actor, doc); err != nil {
				return report, fmt.Errorf("doc %q: %w", doc.Title, err)
			}
			report.Docs++
		}
	}

	return report, nil
}
