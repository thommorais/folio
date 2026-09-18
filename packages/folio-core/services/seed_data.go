package services

import (
	"folio/folio-core/domain"
	"folio/folio-core/ports"
)

// seedSpec is one demo project and everything under it.
type seedSpec struct {
	project ports.CreateProjectInput
	plans   []ports.CreatePlanInput
	tickets []ports.CreateIssueInput
	journal []ports.WriteEntryInput
	docs    []ports.WriteEntryInput
	// done advances the first todos of plan i to the given statuses, so the
	// demo shows plans in progress rather than every one at 0%.
	done [][]domain.IssueStatus
}

// seedProjects is the demo dataset: two projects mid-flight, written the way
// a code agent would leave them. The logs carry real decisions and dead ends,
// since an empty-looking log teaches nothing about what the log is for.
func seedProjects() []seedSpec {
	return []seedSpec{
		{
			project: ports.CreateProjectInput{
				Slug:  "folio",
				Name:  "folio",
				Descr: "Shared workspace an agent writes to: plans, todos, work logs and docs.",
			},
			tickets: []ports.CreateIssueInput{
				{
					Slug:        "search-ranking-ignores-recency",
					Title:       "Search ranking ignores recency",
					Status:      domain.IssueInProgress,
					Priority:    domain.PriorityHigh,
					Tags:        []string{"search", "bug"},
					ExternalRef: "FOLIO-31",
					Body: `Hits come back ordered by created date across kinds, so an
old doc that matches once outranks a log entry written this morning that
matches three times. Needs a score that combines match count with age.`,
				},
			},
			plans: []ports.CreatePlanInput{
				{
					Title:  "Ship full text search",
					Goal:   "One query across logs, docs, todos and plans, scoped to a project.",
					Status: domain.PlanActive,
					Tags:   []string{"search"},
					Todos: []ports.CreateIssueInput{
						{Title: "Decide on the index strategy", Priority: domain.PriorityHigh, Tags: []string{"search"}},
						{Title: "Write the search repository", Priority: domain.PriorityHigh, Tags: []string{"search"}},
						{Title: "Expose the search endpoint", Priority: domain.PriorityMedium},
						{Title: "Benchmark against 100k log entries", Priority: domain.PriorityMedium, Tags: []string{"perf"}},
						{Title: "Paginate results across kinds", Priority: domain.PriorityLow},
					},
				},
				{
					Title:  "MCP adapter",
					Goal:   "Expose the same use cases as MCP tools so Claude can read and write the workspace directly.",
					Status: domain.PlanDraft,
					Tags:   []string{"mcp", "adapters"},
					Todos: []ports.CreateIssueInput{
						{Title: "Map use cases to MCP tool definitions", Priority: domain.PriorityHigh, Tags: []string{"mcp"}},
						{Title: "Decide auth: static token or OAuth", Priority: domain.PriorityHigh, Tags: []string{"mcp", "auth"}},
						{Title: "Write the streamable HTTP transport", Priority: domain.PriorityMedium},
					},
				},
			},
			done: [][]domain.IssueStatus{
				{domain.IssueDone, domain.IssueDone, domain.IssueInProgress},
			},
			journal: []ports.WriteEntryInput{
				{
					Title:       "Chose SQLite FTS5 over a separate search service",
					Branch:      "feat/search",
					ExternalRef: "FOLIO-12",
					Tags:        []string{"decision", "architecture", "search"},
					Body: `## Context

Search has to span logs, docs, todos and plans, scoped to one project, and
stay useful at the volume an agent generates (logs grow without bound).

## Options

- **Meilisearch / Typesense**: best relevance, but a second service to run,
  back up and keep in sync. Too much operational weight for a tool that is
  meant to be one Go binary plus a SQLite file.
- **LIKE over each collection**: no new dependency, but no ranking and a full
  scan per query.
- **SQLite FTS5**: ships inside the database we already have, gives ranking
  and prefix matching, and stays inside the single-file deployment.

## Decision

FTS5. The search adapter runs one query per requested kind and merges newest
first, so each query stays bounded by the caller's limit.

## Consequence

Relevance ranking is weaker than a dedicated engine, and we accept that. If
search quality becomes the complaint, the SearchRepository port is the seam:
swapping in Meilisearch touches one adapter and nothing above it.`,
				},
				{
					Title:  "Search repository landed, benchmarks pending",
					Branch: "feat/search",
					PR:     "14",
					Tags:   []string{"search"},
					Body: `Implemented ` + "`SearchRepository`" + ` against PocketBase: one filtered query
per kind, merged and sorted newest first, paged after the merge.

Snippets are cut on a word boundary at 200 characters so a hit is judgeable
without fetching the record.

Not yet benchmarked. The merge holds kinds x limit rows in memory, which is
fine at the default limit of 50 and wants checking at 500.`,
				},
				{
					Title:  "Collection rules could not reference journ_members at create time",
					Branch: "feat/schema",
					Tags:   []string{"bugfix", "pocketbase"},
					Body: `## Problem

Boot failed with:

    projects: listRule: Invalid rule. Raw error: invalid left operand
    "@collection.journ_members.project" - failed to load collection
    "journ_members"

The project collection's access rules reference ` + "`journ_members`" + `, but the
migration created projects first, so the target did not exist yet.

## Fix

Split the migration in two: create every collection's structure, then apply
the access rules in a second pass. PocketBase resolves ` + "`@collection`" + `
references at save time, so the target only has to exist by then.

## Note

Only caught by actually booting the app. Worth remembering that the schema
code compiles and unit-tests clean while still being unloadable.`,
				},
				{
					Title: "Reworked logs from event stream to work log",
					Tags:  []string{"decision", "domain"},
					Body: `Logs started out as an ops-style event stream: a severity level, a message,
append-only, one row per event.

That was the wrong model. A log here is context on what was built and how,
the decisions behind it, and where the work currently stands. So:

- ` + "`level`" + ` (debug/info/warn/error) removed; tags classify instead.
- ` + "`message`" + ` became ` + "`title`" + ` plus a markdown ` + "`body`" + `.
- Entries are editable, because the state of a piece of work changes.
- Added ` + "`branch`" + `, ` + "`pr`" + ` and ` + "`ticket`" + ` so an entry is reachable from a code
  reference months later.
- Added ` + "`POST /logs/{id}/append`" + `, since the common case is adding to work
  already written up rather than rewriting it.

Dates stay server-assigned: a log is only useful read chronologically if
the timestamps are trustworthy.`,
				},
			},
			docs: []ports.WriteEntryInput{
				{
					Slug:  "architecture",
					Title: "Architecture",
					Tags:  []string{"architecture"},
					Body: `Hexagonal. The domain has no dependencies, ports declare the contracts, and
adapters implement them.

    domain/     models, rules, errors
    ports/      repositories (driven) + use cases (driving)
    services/   use case implementations, permission checks
    adapters/   pb (PocketBase), httpapi (REST), system (clock, ids, logger)

Two rules keep it honest:

1. Adapters map infrastructure types to domain models at the boundary. Nothing
   above ` + "`adapters/pb`" + ` ever sees a ` + "`*core.Record`" + `.
2. Every use case takes a ` + "`ports.Actor`" + ` and authorises the call itself, so a
   new driving adapter cannot skip a permission check by forgetting one.

Adding a client (MCP, CLI, web) means depending on the use case ports and
nothing deeper.`,
				},
				{
					Slug:  "permissions",
					Title: "Permissions",
					Tags:  []string{"architecture", "security"},
					Body: `Every project has members with one of three roles:

| Role | Read | Write content | Manage members |
| --- | --- | --- | --- |
| owner | yes | yes | yes |
| editor | yes | yes | no |
| viewer | yes | no | no |

A project always keeps at least one owner: the last one can be neither removed
nor demoted, or the project would become unadministrable.

A non-member gets **404, not 403**. Confirming that a project exists is itself
a leak. A member who lacks the level gets 403, since they already know it is
there.

Checks live in the service layer so every client inherits them. PocketBase
collection rules enforce the same tenancy for anything hitting
` + "`/api/collections`" + ` directly, as a second line of defence.`,
				},
			},
		},
		{
			project: ports.CreateProjectInput{
				Slug:  "welligence-web",
				Name:  "Welligence Web",
				Descr: "Upstream asset valuation platform. Rails shell plus a TanStack Router SPA.",
			},
			tickets: []ports.CreateIssueInput{
				{
					Slug:        "consent-banner-blocks-first-pageview",
					Title:       "Consent banner blocks the first pageview",
					Status:      domain.IssueOpen,
					Priority:    domain.PriorityMedium,
					Tags:        []string{"analytics"},
					ExternalRef: "XWWP-4501",
					Body: `The banner mounts before the analytics bootstrap, so the
landing pageview is dropped for anyone who has not already consented. Every
later navigation is counted, which is why the drop only shows in session
starts.`,
				},
			},
			plans: []ports.CreatePlanInput{
				{
					Title:  "Make Exploration and Portfolio v2 responsive",
					Goal:   "Both views usable at tablet width without horizontal scroll.",
					Status: domain.PlanActive,
					Tags:   []string{"ux", "responsive"},
					Todos: []ports.CreateIssueInput{
						{Title: "Audit both views at 820x1180", Priority: domain.PriorityHigh, Tags: []string{"ux"}},
						{Title: "Fix filter bar overflow with no scroll affordance", Priority: domain.PriorityHigh, Tags: []string{"ux"}},
						{Title: "Stop KPI cards truncating values on tablet", Priority: domain.PriorityMedium},
						{Title: "Fix the wells chart x-axis label smear", Priority: domain.PriorityMedium, Tags: []string{"charts"}},
						{Title: "Close sticky dropdowns on Escape and on navigation", Priority: domain.PriorityLow},
					},
				},
			},
			done: [][]domain.IssueStatus{
				{domain.IssueDone, domain.IssueInProgress},
			},
			journal: []ports.WriteEntryInput{
				{
					Title:       "GA4 pageview tracking restored on prod",
					Branch:      "release/r378-ga-pageview-fix",
					PR:          "4873",
					ExternalRef: "XWWP-4420",
					Tags:        []string{"analytics", "bugfix"},
					Body: `## Problem

Prod stopped counting GA4 pageviews around June 8.

PR #4484 deleted the ` + "`gtag('config', <id>)`" + ` call that had been firing GA4's
built-in ` + "`page_view`" + ` since the 2023 migration. It added a GTM container
loader fed a GA4 measurement id (not a real ` + "`GTM-`" + ` container), which
initializes nothing. Net: no hit fires. Verified on prod, zero ` + "`collect`" + `
beacons.

Separately pre-existing: the router hook fired ` + "`gtag('event', 'pageview')`" + `,
which GA4 logged as a custom event, never a real pageview. Harmless while
` + "`config`" + ` did the work, but it was never a valid source on its own.

## Fix

- Env-aware GA4 tag in the erb: loads ` + "`gtag/js`" + `, defines the real
  ` + "`window.gtag`" + `, calls ` + "`config`" + ` with ` + "`send_page_view: false`" + `.
- The router hook is now the sole pageview source and emits a standard
  ` + "`page_view`" + ` with ` + "`page_location`" + `/` + "`page_path`" + `/` + "`page_title`" + `.
- Guarded the GTM id so it only renders for a real ` + "`GTM-`" + ` container.

## Verify

Network, filter ` + "`collect`" + `, expect one ` + "`en=page_view`" + ` per navigation.`,
				},
				{
					Title:  "UX audit: full app walkthrough",
					Branch: "XWWP-5010-responsive",
					Tags:   []string{"ux", "audit"},
					Body: `Browser walkthrough of every main page plus global actions, at desktop
1440x900 and tablet 820x1180.

## Top findings

- Portfolio V2 KPI cards never populate (dead "-" skeletons even with filters
  applied). v1 works; the v1/v2 coexistence in nav is confusing.
- Exploration Analytics: wells chart x-axis renders as a black label smear.
- Tablet: all KPI cards truncate values ("9,55..."); filter bars overflow with
  no scroll affordance on three views.
- Sticky filter dropdowns do not close on Escape or on selection, and survive
  navigation and resize.
- No designed 404 page.

## Pattern debt

Three different filter-bar patterns, four variants of an unlabeled "save view"
icon, three detail-surface metaphors. Worth one pass rather than five fixes.`,
				},
			},
			docs: []ports.WriteEntryInput{
				{
					Slug:  "release-process",
					Title: "Release process",
					Tags:  []string{"process"},
					Body: `Releases cut from ` + "`release/rNNN`" + `. A fix aimed at a release branches from
it, never from ` + "`main`" + `, and is cherry-picked forward.

Port a fix to the next release on a fresh branch off that release, not by
merging the previous release branch: merging drags unrelated commits along.`,
				},
			},
		},
	}
}
