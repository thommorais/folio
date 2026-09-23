# folio

The shared channel between developers and their coding agents: tickets, plans,
todos, work logs and docs, scoped per project, reachable from any client.

One Go core (`packages/folio-core`) holds the domain and use cases. PocketBase
(`apps/base`) provides storage, auth and the admin UI, and serves the REST API.
An MCP server, CLI or web app mounts the same use cases rather than
reimplementing the rules.

## Layout

| Path | What it is |
| --- | --- |
| `packages/folio-core` | Domain, ports, services and adapters (Go). |
| `apps/base` | PocketBase backend serving `/api/folio`. |
| `apps/site` | Next.js web app. |
| `apps/docs` | API contracts. |

## Running

```bash
pnpm install
cd apps/base && go run . serve
```

PocketBase comes up on `:8090`, installs the folio collections on boot and
mounts the API at `/api/folio`. Create the first superuser at
`http://127.0.0.1:8090/_/`, then register users through PocketBase's own auth
endpoints.

```bash
cd packages/folio-core && go test ./...
```

## Demo data

```bash
cd apps/base && go run . seed
```

Creates `user@test.com` / `pass@test` and fills the workspace with two projects
mid-flight: a ticket each, plans with real progress, todos in several states,
and work logs carrying actual decisions. It refuses to run against a database that already
has projects (`--force` overrides).

The seed drives the use cases rather than writing records, so it exercises the
same validation and permission checks as any client.

## Model

A **project** has **members** (owner, editor, viewer) and always keeps at
least one owner. Under it sit **plans** (stated intent), **todos** (the
steps), **logs** (what was built, how, and where it stands) and **docs**
(durable knowledge).

A **ticket** is a unit of work large enough to carry its own plans, todos,
logs and docs: those four each hold an optional ticket, so the same record
either hangs off a ticket or sits loose under the project. Deleting a ticket
detaches its contents rather than destroying them.

**Knowledge** sits outside all of that. A tip, a snippet, a fix worth keeping
is rarely about one project, so a note belongs to none: every signed-in user
reads and writes it, and the project it may name records where it was learned
rather than who may see it.

Search spans every kind at once, ranked by relevance through a SQLite FTS5
index, and knowledge answers from inside any project.

Permissions are checked in the service layer, so every client gets the same
rules; PocketBase collection rules enforce the same tenancy for direct REST
access. Knowledge is the deliberate exception at both layers: its only check
is that the caller is signed in.

## Docs

- [API](apps/docs/folio-api.md)
- [Core architecture](packages/folio-core/README.md)
