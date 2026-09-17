// Package folio wires the hexagon together. A driving adapter needs only the
// App returned here: it holds every use case, already bound to the
// PocketBase-backed repositories.
package folio

import (
	"log/slog"

	pbcore "github.com/pocketbase/pocketbase/core"

	"folio/folio-core/adapters/httpapi"
	"folio/folio-core/adapters/pb"
	"folio/folio-core/adapters/system"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

// App is the assembled set of use cases. Driving adapters depend on these
// interfaces, never on the services or repositories behind them.
type App struct {
	Projects ports.ProjectUseCase
	Plans    ports.PlanUseCase
	Issues   ports.IssueUseCase
	Journal  ports.JournalUseCase
	Cycles   ports.CycleUseCase
	WorkLogs ports.WorkLogUseCase
	Docs     ports.DocUseCase
	Search   ports.SearchUseCase
}

// New builds the use cases against a PocketBase app running in-process.
func New(app pbcore.App, logger *slog.Logger) *App {
	clock := system.Clock{}
	ids := system.IDGenerator{}
	log := system.NewLogger(logger)

	projectRepo := pb.NewProjectRepository(app)
	planRepo := pb.NewPlanRepository(app)
	issueRepo := pb.NewIssueRepository(app)
	journalRepo := pb.NewJournalRepository(app)
	docRepo := pb.NewDocRepository(app)
	cycleRepo := pb.NewCycleRepository(app)
	ticketLogRepo := pb.NewTicketLogRepository(app)
	planLogRepo := pb.NewPlanLogRepository(app)
	todoLogRepo := pb.NewTodoLogRepository(app)
	searchRepo := pb.NewSearchRepository(app)

	guard := services.NewProjectGuard(projectRepo)
	issues := services.NewIssueService(issueRepo, planRepo, journalRepo, docRepo, cycleRepo, guard, clock, ids, log)

	return &App{
		Projects: services.NewProjectService(projectRepo, guard, clock, ids, log),
		Plans:    services.NewPlanService(planRepo, issueRepo, issues, guard, clock, ids, log),
		Issues:   issues,
		Journal:  services.NewJournalService(journalRepo, issueRepo, guard, clock, ids, log),
		Cycles:   services.NewCycleService(cycleRepo, issueRepo, guard, clock, ids, log),
		WorkLogs: services.NewWorkLogService(ticketLogRepo, planLogRepo, todoLogRepo, issueRepo, planRepo, cycleRepo, guard, clock, ids, log),
		Docs:     services.NewDocService(docRepo, issueRepo, guard, clock, ids, log),
		Search:   services.NewSearchService(searchRepo, guard),
	}
}

func (a *App) Deps() httpapi.Deps {
	return httpapi.Deps{
		Projects: a.Projects,
		Plans:    a.Plans,
		Issues:   a.Issues,
		Journal:  a.Journal,
		Cycles:   a.Cycles,
		WorkLogs: a.WorkLogs,
		Docs:     a.Docs,
		Search:   a.Search,
	}
}

// Migrate installs the folio collections. It is idempotent and safe to call
// on every boot.
func Migrate(app pbcore.App) error {
	return pb.Register(app)
}
