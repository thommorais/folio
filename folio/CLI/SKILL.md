---
name: folio-cli
description: How to drive the folio CLI — the `folio` command for projects, tickets, plans, todos, cycles, work logs, journal and docs. Use when reading or writing anything in a folio workspace, when a task mentions folio tickets/todos/plans/cycles/worklogs/journal/docs, when resuming or ending a session on a ticket, when running wayfinder or a PDCA cycle against folio, or when `folio` appears in a command.
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

Tickets, todos, plans, journal, docs, tags and search need a project; without
one they fail with an error naming the three ways to give one. Commands that
address a record by its global id need none: `cycle`, `worklog`, `interview`,
`kb`, and `ticket get`, `ticket brief`, `doc get`, `resume` and `stop` given an
id.

```bash
folio use folio --here      # binds this directory and everything below it
eval "$(folio use folio)"   # exports FOLIO_PROJECT for this shell
folio use                   # prints the current selection to stderr
eval "$(folio use --clear)" # unsets the export; --clear --here drops the binding
```

Precedence: `-p/--project`, then `FOLIO_PROJECT`, then the directory binding.
An agent whose every shell call starts fresh loses the export between calls, so
bind with `--here`. `folio use <project>` resolves the reference against the API
first, so a typo fails there rather than on the next command.

## Authentication

```bash
folio login            # prompts for email and password, caches a token
folio config get       # url, project, token (presence only)
folio config path      # where credentials.json lives
folio config set url https://folio.example.com
```

Precedence for url and token: flag, then environment (`FOLIO_URL`, `FOLIO_TOKEN`),
then the cached login. A token is valid only for the host that issued it, so a
URL from a flag or the environment uses the cached token only when it names the
cached host. Default URL is `https://folio.journ.app`.

`--url` accepts a bare host; loopback (`localhost`, `127.0.0.1`, `[::1]`) gets
`http`, anything else `https`.

## Piping and JSON

`--json` is a persistent flag on every command. Use it when a program parses
the output. To read output yourself, take the default table: it costs fewer
tokens than the JSON.

Bare `folio` on a terminal opens the TUI. In a pipe or CI it prints help
instead, so it is safe to call from a script.

Prose bodies come from stdin with `--body -`:

```bash
folio journal write "Cut the 0.4 release" --body - <<'EOF'
Tagged and pushed. The migration ran clean.
EOF
```

`--body -` works on `journal write/update`, `doc create/update`,
`ticket create/update`, `kb add/update` and `worklog write`. The same `-` reads
stdin for `worklog write -`, `ticket resolve --detail -`,
`interview finish --doc -` and `stop <ticket> -`.
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

## Sessions on a ticket

Open every session on a known ticket with `folio resume <ticket>` and end it
with `folio stop <ticket> -`. Resume prints the latest handoff, the work logs
written since, the map's next steps, open children, live plans, doc pointers,
then the body: the one read a session needs. A `drift:` line means the checkout
differs from the handoff; check out the handed-off branch, or tell the user why
not, before working.

`stop` exits 4 on uncommitted changes and lists them. It never commits: get the
commit approved as the user's rules require, commit, rerun. `--allow-dirty`
records the paths instead, for when the user chose to leave work uncommitted.

The handoff note is the next session's whole briefing, so stop before context
runs out. Write it terse, one fact per line:

```bash
folio stop <ticket> - <<'EOF'
State: coverage widget renders; no entitlement lock.
Next: hasFeature gate on product rows (c1).
Trap: subscribedCountries empty in dev seed.
EOF
```

## Reading

`folio ticket brief <id-or-slug>` returns a ticket with its children, plans,
journal and docs in one call, closed children included: the full picture when
`resume` is not enough. Children come back open first, so the next step is the
first row, and each row names its kind; journal entries are the 10 most recent,
`--recent-journal` overrides. When the current cycle is planned by a map, a
`cycle plan` section names the map, how many decisions are open, and the
takeable ones as `next` rows, so the next decision needs no second call.

`folio search` hits tickets, todos, plans, docs, journal, work logs, resolutions,
decisions and knowledge in one call, each hit with a snippet. With a term, hits
rank by relevance; with only `--tags`, newest first. `--all` searches every
project you can read. Reach for it when you do not know where something lives;
reach for `ticket brief` when you already know the ticket.

```bash
folio search "index strategy" --kind journal,doc --limit 5 --json
folio search --tags decision
```

List commands take `--limit` and kind-specific filters such as `--status`,
`--priority`, `--ticket`, `--plan`, `--branch`, `--since` and `--until`; most
also take `--query/-q`, `--tags` and `--offset`, and `<cmd> list --help` names
the exact set. `--status` and `--tags` are comma separated.
Journal, work log, doc and kb lists come back newest first, so `--limit N` is
the newest N.

Tickets and todos are one kind of record and share one status set:
`open,in_progress,blocked,done,cancelled`. Plans keep their own:
`draft,active,done,abandoned`.

`ticket get` and `doc get` accept an id or a slug, but a **slug only resolves
with a project selected** — it is unique within a project, not globally. The
same holds for `ticket brief`, `resume` and `stop`. With a project selected the
argument is tried as a slug first and then as an id, so either works.

## Tickets form a graph

A ticket can sit under another (`--parent`) and can be blocked by others
(`--depends-on`, comma separated ids). `blocked` is derived on read from whether
any blocker is still open, so it is never set by hand; the separate `blocked`
*status* is the one you set yourself. Cycles in either the parent chain or the
dependency graph are refused at write time. `--depends-on` on `ticket update`
replaces the whole set.

`--wayfinder` is one of `map`, `research`, `prototype`, `grilling`, `task`: a
field of its own, not a tag. A ticket with none of them is **work**; `map` is a
map; the other four are **decisions**.

## The PDCA loop

Work runs in cycles: plan, do, check, act, then resolve. A ticket that ships,
gets a bug and comes back opens cycle 2 rather than overwriting cycle 1. Every
cycle command takes the ticket's id, or its slug with a project selected.

```bash
folio cycle open <ticket>                    # cycle N at plan
folio cycle open <ticket> --map "<title>"    # same, planned by a new wayfinder map
folio cycle next <ticket>                    # one phase forward
folio cycle resolve <ticket> "<what happened>"           # add --close on the last round
folio cycle list <ticket>
```

Rules the server holds you to:

- Phases move one step at a time, forward only. `next` after act is refused:
  resolve instead.
- A new cycle opens only once the current one is resolved.
- A cycle planned by a map stays in plan, and cannot be resolved, while any
  decision on the map is open. The refusal names how many are left.
  Leaving plan marks the map done: its job was to clear the way.
- Decisions never run cycles of their own; open the cycle on the work ticket
  the decision serves.
- Closing a ticket needs a resolution on its **current** cycle. A ticket that
  never opened a cycle needs none, though a decision still needs its answer.

A second pass is a new cycle with a new map: `cycle open <ticket> --map` on
cycle 2 files the map beside cycle 1's, under the same work ticket.

## Wayfinding operations

The wayfinder skill asks the tracker for these. In folio:

| Wayfinder | folio |
|---|---|
| Create the map | `folio cycle open <work> --map "<title>"` when it plans a cycle, else `folio ticket create "<title>" --wayfinder map --parent <work>` |
| Map body: Destination, Notes, Not yet specified, Out of scope | `folio ticket update <map> --body -` with the markdown on stdin |
| Decisions so far | derived: `folio ticket brief <map>` ends each answered child's row with `title: answer`; keep no such section in the body |
| Create a ticket | `folio ticket create "<title>" --parent <map> --wayfinder <type> --body -` with `## Question` on stdin |
| Wire blocking (second pass) | `folio ticket update <id> --depends-on <ids>` |
| Frontier | `folio ticket frontier <map>` |
| Claim | `folio ticket update <id> --assignee me` |
| Resolve | `folio ticket resolve <id> "<one-line answer>" --detail -` with the reasoning on stdin |
| Rule out of scope | `folio ticket resolve <id> "<why>" --cancel` |
| Link an asset or research branch | `folio worklog write "<pointer>" --ticket <id>` |
| Resolve a grilling ticket | `folio interview finish <id> "<one-line answer>" --doc -`, after the interview below |

A decision cannot close without an answer, so `ticket resolve` is the only
close. The answer is the gist the map lists; `--detail` becomes a resolution
entry linked from the ticket and found by `folio search --kind decision`.

## Interviews

A grilling ticket (`--wayfinder grilling`) runs as an interview hosted in folio:
the agent posts rounds from the CLI, the user answers on a page, and `finish`
resolves the ticket. The method (frontier per round, durable gates, no filler)
is /grilling and /domain-modeling; this section is the folio mechanics. Every
verb takes the ticket's id, or its slug with a project selected, and acts on the
ticket's active interview. `folio grill` is an alias.

```bash
folio interview start <ticket>    # first line is the page link; resumes an active interview
folio interview show <ticket>     # topic, round, open questions, handled, agent status, link
folio interview pending <ticket>  # Sends past handled, one JSON line each, never blocks
folio interview patch <ticket>    # JSON on stdin: next round, answers, replies, handled
folio interview finish <ticket> "<answer>" --doc -
folio interview list <ticket>     # active and earlier interviews
```

### Turn loop

The agent never waits on the CLI. Every round is also a chat turn.

1. `start`, then `patch` round 1: one to three independent questions. Print the link and end the turn.
2. The user answers on the page, presses Send, and tells you they are done in the chat.
3. Run `pending`. If it prints `nothing sent since handled N`, say so and stop; never invent answers. Each line is the user's own input: act on it without asking to confirm.
4. Send one `patch` carrying everything: the effect of every action, the next round, and `{"agent":{"handled":<last seq read>}}`. The page clears its "sent" state from `handled`, so never publish the round and `handled` in separate patches.

To resume a session, run `show`, then `pending`, and apply all pending Sends in one turn with one patch whose `handled` is the last seq.

### Handling a Send

A Send is `{seq, at, actions[]}`. Apply each question's actions in this order:

- `answer`: set `status: "answered"` and copy `answer` from the action (`kind` accept, option or text, with `option` or `text`).
- `defer`: `status: "deferred"`. `reopen`: `status: "reopened"` and `answer: null`.
- `explore`: set `explore.rows`, one row per option as `{"option":"a","pros":[...],"cons":[...]}`, 2 to 4 pros and 2 to 4 cons each, every item at most 200 characters.
- `thread`: append `{"who":"user","text":...,"at":<the Send's at>}` and then your reply `{"who":"agent","text":...}`. A thread message never answers the question.
- `finish`: see Finishing below, after the other actions.

An answer that changes the recommendation of a question still open gets a new `rec` and `"updated":true`, both fields of that question: `{"id":"q3","updated":true,"rec":{"option":"a","why":"..."}}`.

```bash
folio interview patch <ticket> <<'JSON'
{"questions":[
  {"id":"q1","status":"answered","answer":{"kind":"accept","option":"b"}},
  {"id":"q4","round":2,"title":"Where do edges leave a node","body":"","deps":["q1"],
   "options":[{"k":"a","text":"Bottom to top"},{"k":"b","text":"Left to right"}],
   "rec":{"option":"a","why":"Reads like the frontier order."}}],
 "terms":[{"term":"frontier","def":"decisions with no open blockers","avoid":["queue"]}],
 "agent":{"handled":1}}
JSON
```

Patch rules the server enforces:

- `null` deletes an optional key: `note`, and a question's `deps`, `body`, `options`, `durable`, `updated`, `answer` or `explore`; the other question fields refuse it. Questions merge by `id`, one level deep, each given field replaced whole. Terms merge by `term`. `thread` only appends.
- A new question needs `round`, `title` and `rec`. `rec` is `{option, why}` when there are options and `{text, why}` when there are none. Options are absent or 2 to 4, lettered a to d.
- A round holds at most 3 questions and an interview at most 200. `deps` name existing questions and cannot cycle.
- Unknown keys are refused, and `agent` takes only `handled`. The server sets the agent to working when `pending` returns Sends and back to waiting on a patch that carries `handled` or on `finish`, and stamps times you omit.
- `patch` prints one summary line, such as `round 3: 2 questions added, 1 answered, handled 7`, never the state.

### Finishing

`finish` is refused before the first question and while any question is open, so the user defers what they will not answer. It can start from the page's Finish, which arrives as a `finish` action in a Send, or from the user saying finish in the chat.

Before running it, show the user the proposed one-line answer and the locked decisions in the chat, and run `finish` only once they confirm. The server then writes the doc as the ticket's resolution entry, resolves and closes the ticket, and locks the page in one transaction. A finished interview is read-only; reopening the ticket starts a new one.

The doc is self-contained: summary, Terms with their Avoid lists, Why, Locked decisions with the options rejected and why, Routine choices, Verified facts, Risks, Deferred, Open threads. It must stay under 500000 characters, so summarise routine choices instead of pasting threads. For each deferred question and open thread, ask whether it becomes a decision ticket on the map, fog in the map's Not yet specified, or stays in the doc only.

## Work logs

```bash
folio worklog write "Mapbox rejects feature-state in a filter" --ticket <id>
folio worklog write - --plan <id> <<'EOF'
Longer note from stdin.
EOF
folio worklog list --ticket <id> --cycle <cycle-id>
```

`list` and `write` take exactly one of `--ticket`, `--plan` or `--todo`;
`delete` takes only the entry id. A ticket work log written while a cycle is
open is stamped with that cycle, which is what `list --cycle` filters on.
Deleting the ticket or plan detaches its work logs, the way it detaches plans
and docs.

## Writing

Every body you write (handoff, work log, resolution, journal) is read by a later
agent. Write it terse: one fact per line, ids instead of restated titles.

`update` is a patch: unset flags are left alone, and passing no field at all is
an error rather than a no-op. `create` takes the title as a positional argument.

A write prints the record's id on the first line, then only what the server
decided: `slug:` when a create made one, `blocked` when a dependency is open,
`status:` on `ticket resolve`. Take the first line to chain the id into the
next command.

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

`ticket update` and `todo update` take `--archive` and `--unarchive`, which are
mutually exclusive. An archived issue keeps its status and drops out of every
list, `ticket brief` children, `ticket frontier` and `search`; `ticket list
--archived` and `todo list --archived` show only the archived ones. `ticket get`
still finds it by id or slug. An archived blocker no longer blocks, and restoring it
blocks again. Archiving does not touch an issue's children.

Deletes are not prompted, since an agent cannot answer a prompt.
`ticket delete` detaches its plans, todos, journal and docs; `plan delete`
detaches its todos.
`project delete` destroys everything under the project and refuses to run
without `--yes`.

## Development

From the repo root: `pnpm cli:build`, `pnpm cli:dev`, `pnpm cli:install`
(builds and installs to `~/.local/bin`, override with `PREFIX`). Inside
`apps/cli` the Makefile has `build`, `test`, `check`, `format`, `clean`.
