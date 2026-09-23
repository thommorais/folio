# folio-core

The folio domain: projects, tickets, plans, todos, logs and docs for a code
agent to
write to and search. Hexagonal, so the same use cases back every client.

```
domain/            models, rules, errors (no dependencies)
  rules/           validation, derived fields, transition legality
ports/             repositories + use case interfaces
services/          use case implementations, permission checks
adapters/
  pb/              PocketBase repositories and schema
  httpapi/         REST adapter mounted on PocketBase's router
  system/          clock, ID generator, slog logger
wire.go            assembles the whole thing
```

## Concepts

A **project** is the unit of tenancy and has many **members**, each an owner,
editor or viewer. A project always keeps at least one owner.

A **plan** is the agent's stated intent for a piece of work. A **todo** is a
step, optionally under a plan. Deleting a plan detaches its todos rather than
destroying them.

A **log entry** documents a piece of work: what was built, how, where it
stands and why it was done that way. Titled, searchable, editable, and
anchored to a branch, PR or external tracker key. A **doc** is durable
knowledge meant to be kept current.

A **ticket** is a unit of work carrying its own plans, todos, logs and docs.
Each of those four holds an optional ticket, so a record either hangs off a
ticket or sits directly under the project; a ticket in another project is
refused, so the reference cannot cross a tenancy boundary. A ticket's
progress counts its todos, and deleting one detaches its contents. **Search**
spans logs, docs, todos, plans and tickets.

## Adding a client

Depend on the `ports` use case interfaces, never on `services` or `adapters`:

```go
useCases := folio.New(pbApp, nil)   // wire.go
httpapi.New(httpapi.Deps{Projects: useCases.Projects, ...}).Mount(e)
```

An MCP server or CLI mounts the same way. Every use case method takes a
`ports.Actor` and authorises the call itself, so a new adapter cannot skip a
permission check by forgetting one.

## Rules worth knowing

- A non-member gets `ErrNotFound`, not `ErrForbidden`: confirming a project
  exists is itself a leak.
- `Blocked` and `Progress` are derived on read, never persisted.
- Cancelled todos leave the progress denominator and cannot be reopened.
- Log dates come from the server clock; an edit moves `UpdatedAt` and never
  `CreatedAt`.
- Batch writes are partial: a rejected item is reported by index, the rest go
  through, and the endpoint answers 207.
- Every listing is clamped to `MaxPageSize`.
- Search is an FTS5 index in `adapters/pb`, mirrored by triggers and ranked by
  bm25 with the title weighted above the body. Adding a searchable collection
  means adding one entry to `sources`.
- Tags for entries and issues live in join tables, not in the `tags` column on
  those records, which nothing writes.

## Tests

```bash
go test ./...
```

Business rules are covered by tests against in-memory doubles; the PocketBase
and HTTP adapters are covered by running the app (see `apps/base`).

## Seeding

`services.Seed` writes the demo dataset through the use case ports, so seeded
data passes the same validation and authorisation as a client write. The
content lives in `services/seed_data.go`; `apps/base` exposes it as
`go run . seed`.
