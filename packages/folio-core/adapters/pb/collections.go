// Package pb implements the driven ports against PocketBase, used in-process
// via core.App. Records are mapped to domain models at this boundary: nothing
// above this package ever sees a *core.Record.
package pb

// Collection and field names. They are constants because the migration, the
// repositories and the search adapter must all agree on them; a typo in a
// string literal would surface as an empty result rather than a failure.
const (
	ColProjects = "journ_projects"
	ColMembers  = "journ_members"
	ColPlans    = "journ_plans"
	ColTickets  = "journ_tickets"
	ColTodos    = "journ_todos"
	ColJournal  = "journ_journal"
	ColDocs     = "journ_docs"
	ColCycles   = "journ_cycles"

	ColTicketLogs = "journ_ticket_logs"
	ColPlanLogs   = "journ_plan_logs"
	ColTodoLogs   = "journ_todo_logs"

	// ColUsers is PocketBase's built-in auth collection.
	ColUsers = "users"
)
