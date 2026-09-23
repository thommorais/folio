---
name: folio-cli
description: How to drive the folio CLI — the `folio` command for projects, tickets, plans, todos, cycles, work logs, journal and docs. Use when reading or writing anything in a folio workspace, when a task mentions folio tickets/todos/plans/cycles/worklogs/journal/docs, or when `folio` appears in a command.
---

# folio CLI

`folio` is a Go/cobra CLI over the folio HTTP API. Record kinds under a project:
**tickets** (units of work, sluggable), **plans** (intent), **todos** (steps),
**journal** (what happened, newest first), **docs** (reference, sluggable).
Under a ticket: **cycles** (PDCA rounds) and **work logs**; plans and todos
carry work logs too.

The journal and the work log are not the same thing. The journal is yours,
kept per project, and nothing requires you to write one. A work log hangs off a
single ticket, plan or todo and carries the context needed to resume that work.

Run `folio <command> --help` for the current flag list. This skill covers what
help output does not say.

## Before anything else: select a project

Every read and write except `project` and `config` needs a project. Without one
they fail with `no project: pass --project, or select one with: eval "$(folio use <project>)"`.

```bash
eval "$(folio use folio)"   # exports FOLIO_PROJECT for this shell
folio use                   # prints the current selection to stderr
eval "$(folio use --clear)" # unsets it
```

The selection lives in the environment, never on disk, so two terminals can sit
on different projects. `folio use <project>` resolves the reference against the
API before printing, so a typo fails there rather than on the next command.

Per-command override: `-p/--project <id-or-slug>`.

## Authentication

```bash
folio login            # prompts for email and password, caches a token
folio config get       # url, project, token (presence only)
folio config path      # where credentials.json lives
folio config set url https://folio.example.com
```

Precedence for url and token: flag, then environment (`FOLIO_URL`, `FOLIO_TOKEN`),
then the cached login. The cache is taken as a **pair** — the URL and token are
only read together, because a token is valid only for the host that issued it.
Default URL is `https://folio.journ.app`.

`--url` accepts a bare host; loopback gets `http`, anything else `https`.

## Piping and JSON

`--json` is a persistent flag on every command. Use it whenever you are parsing
output — the default is a `tabwriter` table meant for a human.

Bare `folio` on a terminal opens the TUI. In a pipe or CI it prints help
instead, so it is safe to call from a script.

Prose bodies come from stdin with `--body -`:

```bash
folio journal write "Cut the 0.4 release" --body - <<'EOF'
Tagged and pushed. The migration ran clean.
EOF
```

`--body -` works on `journal write`, `journal update`, `doc create`, `doc update`;
`journal append --section -` adds to an existing entry without rewriting it.

## Tickets and todos are the same record

A ticket and a todo differ only by kind. Either can sit under the other, so a
todo that grows can carry tickets beneath it, and `--ticket <id>` on `todo
create` files the todo under any issue, not only a ticket.

`--size` takes 1, 2, 3, 5 or 8. Paired with priority it ranks work by value per
unit of effort, so a small high-priority ticket outranks a large one. Leave it
off and the issue scores zero and sorts last.

## Tags are a closed vocabulary

Write commands reject any tag outside the known list, with a spelling
suggestion. Two axes, and a record usually carries one of each:

- **context** (where the work lives): api, backend, cli, db, design, docs, frontend, infra, mcp, mobile, tui, web
- **kind** (what sort of work): bug, chore, decision, deploy, dx, perf, refactor, release, security, spike, test

```bash
folio tags            # tally across todos, plans, journal and docs
folio tags --unused   # also list known tags nothing carries yet
```

A tag marked `*` in that table is not in the vocabulary — it predates the list
or came in through the API directly.

`--tags` on a **write** replaces the whole set, it does not merge. Read the
current tags first if you mean to add one.

`todo create` without `--tags` warns on stderr but succeeds: an untagged todo is
findable only by title. Tag todos you create.

## Reading

`folio ticket brief <id-or-slug>` returns a ticket with its children, plans,
journal and docs in one call. This is what to run when opening a session on a
known ticket. Children come back open first, so the next step is the first row,
and each row names its kind; journal entries are the 10 most recent,
`--recent-journal` overrides.

`folio search` hits journal, docs, todos and plans in one call, newest first, each
hit with a snippet. Reach for it when you do not know where something lives;
reach for `ticket brief` when you already know the ticket.

```bash
folio search "index strategy" --kind journal,doc --limit 5 --json
folio search --tags decision
```

List commands take `--query/-q`, `--tags`, `--limit`, `--offset`, and
kind-specific filters (`--status`, `--priority`, `--ticket`, `--plan`,
`--branch`, `--since`, `--until`). `--status` and `--tags` are comma separated.

Tickets and todos are one kind of record and share one status set:
`open,in_progress,blocked,done,cancelled`. Plans keep their own:
`draft,active,done,abandoned`.

`ticket get` and `doc get` accept an id or a slug, but a **slug only resolves
with a project selected** — it is unique within a project, not globally. The
same holds for `ticket brief`. With a project selected the argument is tried as
a slug first and then as an id, so either works.

## Tickets form a graph

A ticket can sit under another (`--parent`) and can be blocked by others
(`--depends-on`, comma separated ids). `blocked` is derived on read from whether
any blocker is still open, so it is never set by hand; the separate `blocked`
*status* is the one you set yourself. Cycles in either the parent chain or the
dependency graph are refused at write time.

```bash
folio ticket create "The map" --wayfinder map
folio ticket create "Decide the shape" --parent <map-id> --wayfinder grilling
folio ticket create "Build it" --parent <map-id> --depends-on <id1>,<id2>
folio ticket frontier <map-id>
```

`ticket frontier <id>` lists that ticket's children that are open, unblocked and
unassigned, oldest first. `--wayfinder` is one of `map`, `research`,
`prototype`, `grilling`, `task`, and is a field of its own rather than a tag, so
it does not touch the tag vocabulary.

## Cycles carry the PDCA loop

A ticket that ships, gets a bug and comes back opens a second cycle rather than
overwriting the first. Each cycle has an ordinal, a phase and a resolution, and
`ticket get` reports the latest as `cycle N: phase`.

```bash
folio cycle open <ticket-id>          # starts at plan
folio cycle phase <cycle-id> do       # plan -> do -> check -> act, one step
folio cycle resolve <cycle-id> "Shipped behind a flag"
folio cycle list <ticket-id>
```

Phases advance one step at a time and never go backwards. A new cycle cannot
open while the current one is unresolved, and **closing a ticket requires a
resolution on its current cycle** — an earlier cycle's resolution does not
count. A ticket that never opened a cycle closes freely, so this binds only
work that opted into the loop.

## Work logs

```bash
folio worklog write "Mapbox rejects feature-state in a filter" --ticket <id>
folio worklog write - --plan <id> <<'EOF'
Longer note from stdin.
EOF
folio worklog list --ticket <id> --cycle <cycle-id>
```

Exactly one of `--ticket`, `--plan` or `--todo` is required on every subcommand.
A ticket work log written while a cycle is open is **stamped with that cycle**
automatically, which is what `--cycle` then filters on. Work logs cascade with
their parent rather than detaching the way plans and docs do.

## Writing

`update` is a patch: unset flags are left alone, and passing no field at all is
an error rather than a no-op. `create` takes the title as a positional argument.

```bash
folio ticket create "Mobile nav" --body "No nav below md." --tags frontend,bug
folio todo create "Add the hamburger" --ticket <id> --tags frontend,bug
folio todo done <id>
folio journal write "Shipped mobile nav" --branch develop --pr 42 --ticket <id> --tags frontend,release
```

`todo start`, `todo done` and `todo cancel` are shortcuts over the same patch
`--status` writes, so either spelling does the same thing. `--status` stays the
way to reach the statuses without a shortcut, `open` and `blocked`.

`todo block <id> --on <ids>` is the exception: it records a dependency rather
than writing the blocked status, since `blocked` on a read is derived from
`depends_on` and never persisted. It appends to the set rather than replacing
it, `--off` drops a blocker, and the two flags are mutually exclusive. The
explicit status is still `todo update <id> --status blocked`, and the two are
independent: a todo can carry the status without a dependency, or derive
blocked without the status.

Deletes cascade downward and are not prompted, since an agent cannot answer a
prompt. `ticket delete` detaches its plans, todos, journal and docs.
`project delete` destroys everything under the project and refuses to run
without `--yes`.

## Development

From the repo root: `pnpm cli:build`, `pnpm cli:dev`, `pnpm cli:install`
(builds and installs to `~/.local/bin`, override with `PREFIX`). Inside
`apps/cli` the Makefile has `build`, `test`, `check`, `format`, `clean`.
