# folio API

REST, mounted on PocketBase at `/api/folio`. Auth is PocketBase's: send a user
token as the `Authorization` header. Anonymous requests get 401.

```bash
TOKEN=$(curl -s -X POST localhost:8090/api/collections/users/auth-with-password \
  -H 'Content-Type: application/json' \
  -d '{"identity":"you@example.com","password":"..."}' | jq -r .token)
```

`{project}` accepts either the project id or its slug.

## Status codes

| Code | Meaning |
| --- | --- |
| 400 | validation failed; body carries `field` and `reason` |
| 401 | no or invalid token |
| 403 | a member without the required role |
| 404 | absent, or a project the caller is not a member of |
| 409 | slug already taken |
| 207 | batch write partly succeeded; body lists `created` and `errors` |

## Projects

| Method | Path | Role |
| --- | --- | --- |
| GET | `/projects?archived=true` | member |
| POST | `/projects` | any user, becomes owner |
| GET | `/projects/{project}` | member |
| PATCH | `/projects/{project}` | editor |
| DELETE | `/projects/{project}` | owner |
| POST | `/projects/{project}/members` | owner |
| PATCH | `/projects/{project}/members/{user}` | owner |
| DELETE | `/projects/{project}/members/{user}` | owner |

`POST /projects` derives the slug from the name when one is not supplied.
Members are added by email, and the last owner can be neither removed nor
demoted.

```bash
curl -X POST localhost:8090/api/folio/projects -H "Authorization: $TOKEN" \
  -d '{"name":"Search Rewrite"}'          # -> slug "search-rewrite"

curl -X POST localhost:8090/api/folio/projects/search-rewrite/members \
  -H "Authorization: $TOKEN" -d '{"email":"bob@example.com","role":"editor"}'
```

## Tickets

A ticket is a unit of work under a project: a bug, a feature, an
investigation. Plans, todos, journal and docs each carry an optional `ticket_id`,
so the same four kinds of record hang off a ticket or sit loose under the
project.

| Method | Path | Role |
| --- | --- | --- |
| GET | `/projects/{project}/tickets?status=open,in_progress` | member |
| POST | `/projects/{project}/tickets` | editor |
| GET | `/projects/{project}/tickets/{slug}` | member |
| GET | `/projects/{project}/tickets/{slug}/brief` | member |
| GET | `/tickets/{ticket}` | member |
| GET | `/tickets/{ticket}/brief` | member |
| PATCH | `/tickets/{ticket}` | editor |
| DELETE | `/tickets/{ticket}` | editor |

Filters: `status`, `priority`, `assignee`, `tags`, `q`, `limit`, `offset`.
Status: `open`, `in_progress`, `blocked`, `closed`, `cancelled`. Priority:
`low`, `medium`, `high`.

The slug is derived from the title when omitted, and is unique per project.
Every read reports `progress`, counted from the ticket's todos the same way a
plan's is. `external_ref` names the same work in another tracker.

A closed ticket reopens; a cancelled one does not. Deleting a ticket detaches
its plans, todos, journal and docs rather than deleting them: they fall back to
the project.

```bash
curl -X POST localhost:8090/api/folio/projects/search-rewrite/tickets \
  -H "Authorization: $TOKEN" -d '{
    "title":"Search ranking ignores recency","priority":"high",
    "external_ref":"FOLIO-31","tags":["search","bug"]}'   # -> slug "search-ranking-ignores-recency"

curl -X PATCH localhost:8090/api/folio/tickets/$ID -H "Authorization: $TOKEN" \
  -d '{"status":"in_progress"}'
```

`brief` returns the ticket with everything filed under it in one response:
`ticket`, `plans`, `todos`, `journal`, `docs`. Todos come back open first, so the
next step is the first row. Journal entries are capped at the 10 most recent; `recent_journal`
overrides that. Empty collections are `[]`, never `null`.

```bash
curl localhost:8090/api/folio/tickets/$ID/brief?recent_journal=3 -H "Authorization: $TOKEN"
```

Filing a child under a ticket in another project is rejected, so a ticket
reference cannot pull a record across a project boundary.

## Plans

| Method | Path | Role |
| --- | --- | --- |
| GET | `/projects/{project}/plans?status=active,draft` | member |
| POST | `/projects/{project}/plans` | editor |
| GET | `/plans/{plan}` | member |
| PATCH | `/plans/{plan}` | editor |
| DELETE | `/plans/{plan}` | editor |

A plan carries its todos in the same call, and every read reports `progress`
(cancelled todos excluded). Deleting a plan detaches its todos. A plan filed
under a ticket passes its `ticket_id` down to the todos created with it, so
the ticket's progress counts work planned under it.

```bash
curl -X POST localhost:8090/api/folio/projects/search-rewrite/plans \
  -H "Authorization: $TOKEN" -d '{
    "title":"Ship full text search","status":"active",
    "todos":[{"title":"design the index"},{"title":"write the adapter","priority":"high"}]}'
```

Status: `draft`, `active`, `done`, `abandoned`.

## Todos

| Method | Path | Role |
| --- | --- | --- |
| GET | `/projects/{project}/todos` | member |
| POST | `/projects/{project}/todos` | editor |
| GET | `/todos/{todo}` | member |
| PATCH | `/todos/{todo}` | editor |
| DELETE | `/todos/{todo}` | editor |

Filters: `plan_id`, `ticket_id`, `status`, `priority`, `tags`, `q`, `limit`,
`offset`.
Status: `pending`, `in_progress`, `done`, `blocked`, `cancelled`. Priority:
`low`, `medium`, `high`.

POST takes one todo, or a batch under `items`. `blocked` is derived from
`depends_on`: true while any dependency is still open. A cancelled todo cannot
be reopened.

```bash
curl -X POST localhost:8090/api/folio/projects/search-rewrite/todos \
  -H "Authorization: $TOKEN" \
  -d '{"items":[{"title":"write the docs"},{"title":"deploy"}]}'

curl -X PATCH localhost:8090/api/folio/todos/$ID -H "Authorization: $TOKEN" \
  -d '{"status":"done"}'
```

## Logs

The work log: titled, searchable entries documenting what was built, how, and
where it stands. Entries are editable, since the state of a piece of work
changes as it progresses, and dated by the server so a project reads back
chronologically.

| Method | Path | Role |
| --- | --- | --- |
| GET | `/projects/{project}/journal` | member |
| POST | `/projects/{project}/journal` | editor |
| GET | `/journal/{entry}` | member |
| PATCH | `/journal/{entry}` | editor |
| POST | `/journal/{entry}/append` | editor |
| DELETE | `/journal/{entry}` | editor |

Only `title` is required; an entry may start as a stub and be filled in as the
work proceeds. `body` is markdown. `branch`, `pr` and `external_ref` anchor the
entry to the work, and an agent can fill them from git for free. `external_ref`
is a key in another tracker and is distinct from `ticket_id`, which points at a
folio ticket.

```bash
curl -X POST localhost:8090/api/folio/projects/welligence-web/journal \
  -H "Authorization: $TOKEN" -d '{
    "title":"GA4 pageview tracking restored on prod",
    "body":"## Problem\nPR #4484 deleted the gtag config call...",
    "branch":"release/r378-ga-pageview-fix","pr":"4873","external_ref":"XWWP-4420",
    "tags":["analytics","decision"]}'
```

`POST /journal/{entry}/append` adds a `section` to the body, separated by a blank
line, so recording later progress on the same work does not mean reading and
resending the whole entry.

```bash
curl -X POST localhost:8090/api/folio/journal/$ID/append -H "Authorization: $TOKEN" \
  -d '{"section":"## Ported to r379\nClean cherry-pick, no conflicts."}'
```

Filters: `q` (title and body), `branch`, `external_ref`, `ticket_id`,
`plan_id`, `todo_id`, `tags`, `since`, `until` (RFC 3339), `limit`, `offset`.
Newest first.

`created_at` and `updated_at` come from the server clock, never the caller.
Editing an entry moves `updated_at` and leaves `created_at` alone.

## Docs

| Method | Path | Role |
| --- | --- | --- |
| GET | `/projects/{project}/docs` | member |
| POST | `/projects/{project}/docs` | editor |
| GET | `/projects/{project}/docs/{slug}` | member |
| GET | `/docs/{doc}` | member |
| PATCH | `/docs/{doc}` | editor |
| DELETE | `/docs/{doc}` | editor |

Filters: `ticket_id`, `tags`, `q`, `limit`, `offset`. The slug is derived from
the title when omitted, and is unique per project.

## Search

`GET /projects/{project}/search?q=FTS5&kind=doc,log&tags=&limit=&offset=`

Spans `log`, `doc`, `todo`, `plan` and `ticket`; omit `kind` for all five. Log hits
match on title and body, so a decision recorded weeks ago is findable by a
phrase from it. Results are newest first, each with a `snippet`. A query with
neither `q` nor `tags` is rejected: it would scan the project.
