// Package pb implements the driven ports against PocketBase, used in-process
// via core.App. Records are mapped to domain models at this boundary: nothing
// above this package ever sees a *core.Record.
package pb

// Collection and field names. They are constants because the migration, the
// repositories and the search adapter must all agree on them; a typo in a
// string literal would surface as an empty result rather than a failure.
const (
	ColClients   = "journ_clients"
	ColDomains   = "journ_domains"
	ColProjects  = "journ_projects"
	ColIssues    = "journ_issues"
	ColLinks     = "journ_issue_links"
	ColEntries   = "journ_entries"
	ColTags      = "journ_tags"
	ColIssueTags = "journ_issue_tags"
	ColEntryTags = "journ_entry_tags"
	ColMembers   = "journ_members"
	ColPlans     = "journ_plans"
	ColCycles    = "journ_cycles"

	// ColUsers is PocketBase's built-in auth collection.
	ColUsers = "users"
)
