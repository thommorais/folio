# PocketBase realtime: concurrent subscriptions and their cost

Researched 2026-09-15.

Client: `pocketbase` JS SDK **0.27.3** (declared `^0.27.1` in `apps/web/package.json`, resolved in `pnpm-lock.yaml`, installed at `node_modules/.pnpm/pocketbase@0.27.3/node_modules/pocketbase`).

Server: `github.com/pocketbase/pocketbase` **v0.40.3** (pinned in `apps/base/go.mod` and `packages/folio-core/go.mod`).

## Source provenance

The shipped SDK bundle has no readable original source, but `dist/pocketbase.es.mjs.map` embeds the full pre-bundle TypeScript in `sourcesContent`. The extracted `src/services/RealtimeService.ts` is byte-identical (18690 bytes) to `https://raw.githubusercontent.com/pocketbase/js-sdk/v0.27.3/src/services/RealtimeService.ts`, so the line numbers below refer to that file and describe the code actually installed.

The Go source was read from the local module cache at `/Users/thommorais/go/pkg/mod/github.com/pocketbase/pocketbase@v0.40.3`, verified byte-identical to `https://raw.githubusercontent.com/pocketbase/pocketbase/v0.40.3/apis/realtime.go`.

## 1. One SSE connection, multiplexed

All topics share a single `EventSource`. There is exactly one `eventSource` field on the service and it is created in one place.

`RealtimeService.ts:17` declares `private eventSource: EventSource | null = null`, and `RealtimeService.ts:419` is the only construction site:

```ts
this.eventSource = new EventSource(this.client.buildURL("/api/realtime"));
```

Each topic is a named SSE event on that one stream. `subscribe()` registers a DOM listener keyed by the topic string (`RealtimeService.ts:109`, `RealtimeService.ts:377`: `this.eventSource.addEventListener(key, listener)`). Adding a subscription never opens a second connection. `connect()` is only reached when `isConnected` is false (`RealtimeService.ts:101-103`), and `isConnected` is true once an `eventSource` and a `clientId` exist (`RealtimeService.ts:35-37`).

## 2. Each `subscribe()` queues a POST, but concurrent calls coalesce into one

There is no `setTimeout` debounce. The batching is a microtask queue, which collapses any number of `subscribe()` calls made in the same synchronous block (or same microtask turn) into a single `POST /api/realtime`.

`submitSubscriptions` (`RealtimeService.ts:244-252`) pushes a promise pair onto `pendingSubmits` and only schedules the flush when it is the first queued entry:

```ts
private async submitSubscriptions(): Promise<void> {
    return new Promise((resolve, reject) => {
        this.pendingSubmits.push({ resolve, reject });

        if (this.pendingSubmits.length == 1) {
            queueMicrotask(() => this.finalizePendingSubscriptions());
        }
    });
}
```

`finalizePendingSubscriptions` (`RealtimeService.ts:254-283`) drains the whole queue, makes one `sendSubscriptions()` call, and resolves every queued promise from that single request. Its `finally` block re-runs itself if more submits arrived while the request was in flight, so overlapping bursts serialise rather than fan out.

`sendSubscriptions` (`RealtimeService.ts:331-366`) sends the **full topic set**, not a delta:

```ts
return this.client
    .send("/api/realtime", {
        method: "POST",
        body: {
            clientId: this.clientId,
            subscriptions: this.lastSentSubscriptions,
        },
        requestKey: this.getSubscriptionsCancelKey(),
    })
```

Two further suppressions matter. `hasUnsentSubscriptions()` (`RealtimeService.ts:316-329`) compares the current topic set against `lastSentSubscriptions` and returns early with no HTTP request at all if the set is unchanged (`RealtimeService.ts:343-345`). And `requestKey` is a constant per client (`"realtime_" + this.clientId`, `RealtimeService.ts:285-287`), so the SDK's auto-cancellation aborts any earlier in-flight subscription POST when a newer one starts; the abort is swallowed at `RealtimeService.ts:360-365`.

The server treats each POST as a full replace, matching the SDK sending the complete set. `apis/realtime.go:233-237`:

```go
// unsubscribe from any previous existing subscriptions
e.Client.Unsubscribe()

// subscribe to the new subscriptions
e.Client.Subscribe(e.Subscriptions...)
```

The docs state the same: "Sets new active client's subscriptions (and auto unsubscribes from the previous ones)" (https://pocketbase.io/docs/api-realtime/).

For folio this means the seven `ticket/index.tsx` subscriptions and the five in `use-counts.ts` do not produce twelve POSTs. Each hook's `subscribeToList` is an independent async call, so they land in one or a small number of microtask turns; worst case is one POST per turn, each carrying the complete topic set, with unchanged-set calls suppressed entirely.

## 3. Hard server cap of 1000 subscriptions per client; no documented limit

The server validates the incoming subscription list. `apis/realtime.go:173-181`:

```go
func (form *realtimeSubscribeForm) validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.ClientId, validation.Required, validation.Length(1, 255)),
		validation.Field(&form.Subscriptions,
			validation.Length(0, 1000),
			validation.Each(validation.Length(0, 2500)),
		),
	)
}
```

So: at most **1000 topics** per client, each topic string at most **2500 characters**. Exceeding either returns 400. The 2500-character ceiling is the one folio could plausibly approach, because the serialised `filter` is part of the topic string (see section 4), not the subscription count.

The SDK imposes no cap of its own; `subscriptions` is an unbounded object (`RealtimeService.ts:18`).

The official docs do not document either limit. The page at https://pocketbase.io/docs/api-realtime/ describes the endpoints and the replace semantics but states no maximum. The 1000/2500 figures above come only from the server source.

Two related server-side limits, not subscription counts: the SSE connection has a 30-minute maximum lifetime (`apis/realtime.go:70`, `connectEvent.MaxTimeout = 30 * time.Minute`), and the docs state the server sends a disconnect signal if a client receives no message for 5 minutes (https://pocketbase.io/docs/api-realtime/).

There is no documented or source-visible cap on the number of realtime *clients* per server, nor a per-IP subscription limit. Not determinable from the sources: any practical throughput ceiling in records/second, which would depend on deployment. Checked `apis/realtime.go`, `tools/subscriptions/client.go`, `tools/subscriptions/broker.go` and the realtime docs page.

## 4. Filters are encoded into the topic string, and cost one SQL query per event per subscription

**Encoding.** `RecordService.subscribe` prefixes the collection (`RecordService.ts:159-163`), producing `collection/topic`. `RealtimeService.subscribe` then appends the options as a single URL-encoded JSON blob (`RealtimeService.ts:70-82`):

```ts
let key = topic;

if (options) {
    options = Object.assign({}, options); // shallow copy
    normalizeUnknownQueryParams(options);
    const serialized =
        "options=" +
        encodeURIComponent(
            JSON.stringify({ query: options.query, headers: options.headers }),
        );
    key += (key.includes("?") ? "&" : "?") + serialized;
}
```

`normalizeUnknownQueryParams` (`tools/options.ts:135-149`) moves any key not in `knownSendOptionsKeys` into `options.query`. `filter` and `sort` are not in that list (`tools/options.ts:110-132`), so they become query params. A folio call like `collection().subscribe('*', fn, { filter: "..." })` therefore yields a topic key of the shape:

```
tickets/*?options=%7B%22query%22%3A%7B%22filter%22%3A%22...%22%7D%7D
```

**This topic key is the subscription identity.** Two subscriptions differing only in filter text are different topics, both to the SDK and to the server. That is also why a long filter eats into the 2500-character per-topic budget.

**Server-side evaluation.** The server parses the `options` param back out when registering the subscription (`tools/subscriptions/client.go:163-194`), storing `Query` and `Headers` alongside the topic.

On each record change, `realtimeBroadcastRecord` builds a prefix map of the six topic shapes (`apis/realtime.go:605-613`) and iterates every matching subscription of every client (`apis/realtime.go:632-641`), calling `realtimeCanAccessRecord` per subscription.

The filter is **not** an in-memory expression evaluation. `realtimeCanAccessRecord` (`apis/realtime.go:856-901`) compiles the filter to SQL and runs a real query against the database, scoped to the single changed record id:

```go
filter := requestInfo.Query[search.FilterQueryParam]
if filter == "" {
    return true // no further checks needed
}
...
var exists int

q := app.ConcurrentDB().Select("(1)").
    From(record.Collection().Name).
    AndWhere(dbx.HashExp{record.Collection().Name + ".id": record.Id})

resolver := core.NewRecordFieldResolver(app, record.Collection(), requestInfo, false)
expr, err := search.FilterData(filter).BuildExpr(resolver)
...
err = q.Limit(1).Row(&exists)

return err == nil && exists > 0
```

So the cost per record change is: (number of distinct matching subscriptions across all connected clients) x (one collection access-rule check + one `SELECT (1) ... WHERE id = ? AND <filter> LIMIT 1`). It is an indexed single-row lookup, so each query is cheap, but it is a database round trip and it scales linearly with the number of distinct filtered subscriptions, not with the number of listeners. The comment at `apis/realtime.go:868` calls it "the subscription client-side filter", but the check is executed server-side.

Note this happens before any access-rule short-circuit benefit: the collection rule check (`app.CanAccessRecord`) runs first at `apis/realtime.go:864`, and a failing rule skips the filter query.

## 5. Reconnect resubscribes automatically; events during the gap are lost

**Resubscription is automatic.** On connection failure, `connectErrorHandler` (`RealtimeService.ts:490-519`) schedules a retry with a backoff drawn from `predefinedReconnectIntervals` (`RealtimeService.ts:25-27`):

```ts
private predefinedReconnectIntervals: Array<number> = [
    200, 300, 500, 1000, 1200, 1500, 2000,
];
```

It indexes by `reconnectAttempts` and clamps to the last element (`RealtimeService.ts:510-514`), so backoff rises 200ms to 2000ms and then stays at 2000ms forever. `maxReconnectAttempts` is `Infinity` (`RealtimeService.ts:24`), so the SDK retries indefinitely and never gives up on its own.

Each retry calls `initConnect()`, which opens a fresh `EventSource`. The server issues a **new** `clientId` on the new stream, and the `PB_CONNECT` handler (`RealtimeService.ts:427-432`) captures it and deliberately clears the sent-set so the full topic list is re-POSTed:

```ts
this.eventSource.addEventListener("PB_CONNECT", (e) => {
    const msgEvent = e as MessageEvent;
    this.clientId = msgEvent?.lastEventId;
    this.lastSentSubscriptions = [];

    this.submitSubscriptions()
```

Resetting `lastSentSubscriptions = []` is what defeats the no-change suppression from section 2 and forces a real POST. The handler then re-checks up to three more times (`RealtimeService.ts:450-454`) and once again after resolving (`RealtimeService.ts:478-480`) to catch topics added while the connect was in flight. The server comment agrees: "in case of reconnect, clients will have to resubmit all subscriptions again" (`apis/realtime.go:183`).

Note that during reconnection `connect()` returns immediately rather than blocking (`RealtimeService.ts:394-399`), so a `subscribe()` call issued mid-reconnect resolves before the topic has reached the server; the topic is still in `this.subscriptions` and gets picked up by the next `PB_CONNECT` flush.

**Events in the gap are lost.** There is no replay. Nothing in the SDK sends a `Last-Event-ID` header or any cursor, and nothing in the server buffers per-client messages. The broker holds only a live subscription set (`tools/subscriptions/client.go`); `realtimeBroadcastRecord` fans out to currently-connected clients and drops the message for anyone absent. `disconnect()` (`RealtimeService.ts:521-548`) tears down the `EventSource` and clears `clientId`, and the new connection starts with a new `clientId` and an empty history. The docs describe no replay mechanism.

**Conclusion: a refetch-on-reconnect is required.** Any create/update/delete that lands between the drop and the successful resubscribe will never be delivered, so client state silently diverges. The SDK provides the two hooks needed to do this: `onDisconnect` (`RealtimeService.ts:51`), whose argument distinguishes an intentional teardown (`activeSubscriptions.length == 0`) from a network fault, and the `PB_CONNECT` topic, which can be subscribed to like any other and fires on every successful connect including reconnects (`RealtimeService.ts:468-473`). Because `PB_CONNECT` also fires on the very first connect, a refetch handler attached to it will run once redundantly at startup.

## 6. Two handlers on the same topic cost one server-side subscription

One. Deduplication happens twice over, on both sides.

**Client.** `subscriptions` is `{ [key: string]: Array<EventListener> }` (`RealtimeService.ts:10`, `RealtimeService.ts:18`), keyed by the topic string including the serialised options. A second `subscribe()` with the same key pushes onto the existing array (`RealtimeService.ts:96-99`) rather than creating a new entry. The topic list sent to the server is derived from the object's keys, so duplicates cannot appear in it. `getNonEmptySubscriptionKeys` (`RealtimeService.ts:304-314`):

```ts
for (let key in this.subscriptions) {
    if (this.subscriptions[key].length) {
        result.push(key);
    }
}
```

Better still, the second subscribe usually sends no HTTP request at all. `RealtimeService.ts:104-110` only calls `submitSubscriptions()` when the listener is the first for that key:

```ts
} else if (this.subscriptions[key].length === 1) {
    // send the updated subscriptions (if it is the first for the key)
    await this.submitSubscriptions();
} else {
    // only register the listener
    this.eventSource?.addEventListener(key, listener);
}
```

The second handler is attached as another DOM listener on the same named SSE event and costs nothing on the wire. Even if a POST were somehow triggered, `hasUnsentSubscriptions()` would find the set unchanged and skip it.

**Server.** `c.subscriptions` is a map keyed by the topic string (`tools/subscriptions/client.go:194`, `c.subscriptions[s] = options`), so a repeated topic in the POST body collapses to one entry. One entry means one access check and one filter query per event, regardless of how many client-side handlers are attached.

The caveat is that the key includes the serialised options. Identical topics with **different** filter strings are distinct subscriptions and each pays its own query in section 4. The strings must match exactly, including key order inside the JSON, since the key is the output of `JSON.stringify` over the options object.

## What this means for folio

**Deduplication is not needed.** The SDK already does it, at a finer grain than application code could. Twelve `subscribe()` calls across `pages/ticket/index.tsx` (seven) and `app/use-counts.ts` (five) produce one SSE connection, and the microtask batching in `submitSubscriptions` collapses same-turn calls into a single `POST /api/realtime`. Adding one `collection.subscribe(recordId, ...)` per open single-record view adds one topic to a set capped at 1000; at folio's scale that is not a concern. Building an application-level subscription registry would duplicate `RealtimeService.subscriptions` and risk diverging from it.

Two things are worth watching, neither of which is deduplication:

The real cost is **distinct filter strings**, not subscription count. Every distinct filtered topic costs one `SELECT ... WHERE id = ? AND <filter> LIMIT 1` per matching record change, per connected client (section 4). The five `use-counts.ts` subscriptions call `subscribeToList(project, recount)` with no filter argument, so they hit the `filter == ""` fast path at `apis/realtime.go:871` and cost only the collection rule check. The `ticket/index.tsx` hooks all pass `{ ticketId }`, which becomes a filter, so each is a genuine per-event query. Since dedup is by exact topic string, `columns()` and `client.filter()` should produce byte-stable output for equal inputs; unstable parameter ordering would silently split one subscription into several, each paying its own query. The planned per-record subscriptions are unfiltered single-record topics, which are the cheapest kind.

**A refetch-on-reconnect is needed.** This is the actionable finding. The SDK restores the subscription set automatically and retries forever, but events during the gap are gone with no replay (section 5). With backoff clamped at 2000ms, a flaky connection leaves repeated multi-second windows in which writes are missed and the UI shows stale data with no error. Attach a refetch to the `PB_CONNECT` topic, which fires on every successful connect; guard against the redundant first-connect run, or use `onDisconnect` with `activeSubscriptions.length > 0` to arm the refetch only after a genuine fault. Given the 30-minute connection lifetime cap (`apis/realtime.go:70`), every long-lived session will reconnect at least once regardless of network quality, so this path is routine rather than exceptional.
