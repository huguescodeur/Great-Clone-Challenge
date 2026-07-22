# Architecture

Design decisions, trade-offs, and known compromises for insta-lite.

---

## Table of contents

1. [Overview](#overview)
2. [Request lifecycle](#request-lifecycle)
3. [Feed — hybrid push/pull fanout](#feed--hybrid-pushpull-fanout)
4. [Live counters](#live-counters)
5. [Notifications — async event workers](#notifications--async-event-workers)
6. [Rate limiting](#rate-limiting)
7. [Auth — JWT + Redis blacklist](#auth--jwt--redis-blacklist)
8. [Graceful shutdown](#graceful-shutdown)
9. [Database migrations](#database-migrations)
10. [Redis resilience — fail-open](#redis-resilience--fail-open)
11. [Known limitations and conscious trade-offs](#known-limitations-and-conscious-trade-offs)

---

## Overview

```
Client
  │
  ▼
chi router + middleware (logger, rate limiter, JWT auth)
  │
  ▼
Handler → Service → Store (PostgreSQL via pgx)
                 └──→ Redis (cache, counters, blacklist, rate limit)
                 └──→ Event bus (in-process channel)
                            │
                            ▼
                       Worker goroutines
                         ├── Fanout (feed cache)
                         └── Notification writers
```

Each domain (`post`, `like`, `comment`, `commentlike`, `follow`, `feed`, `notification`) is a self-contained package with its own handler, service, store, and model. The `app` package is the only place that imports and wires them all together.

---

## Request lifecycle

1. `chi` router matches the path and runs middleware in order: logger → global rate limiter → JWT auth (if protected).
2. The handler extracts validated input and calls the service.
3. The service owns the business rules, runs DB operations (possibly in a transaction), and optionally publishes an event to an in-process bus.
4. The handler serialises the response to JSON.

The service layer never writes directly to HTTP — that stays in the handler. The store layer never knows about business rules — that stays in the service.

---

## Feed — hybrid push/pull fanout

### Why not a simple `JOIN` on follows?

A `SELECT posts JOIN followers` query works at small scale but degrades as follower counts grow. A user with 10 000 followers reading their feed would cause a full join scan on every request.

### Push (write fanout) — regular users

When a regular user creates a post, a `PostCreatedEvent` is published to an async bus. A fanout worker picks it up and pushes the post ID into each follower's Redis sorted set:

```
Key:   feed:{followerID}
Score: UnixNano(post.CreatedAt)   ← sorts newest-first
Value: postID
```

Reading the feed is then O(log N) on a sorted set, with no DB join.

**Trade-off**: write amplification. A user with 500 followers causes 500 Redis writes per post. Acceptable for regular accounts, unacceptable for celebrities.

### Pull (read fanout) — high-follower accounts

When the author's `followers_count` exceeds `pushThreshold` (currently 1000), the post is written once into a separate sorted set:

```
Key:   celebrity_feed:{authorID}
Score: UnixNano(post.CreatedAt)
Value: postID
```

On read, `FeedService.getCelebrityEntries` checks which followed accounts are celebrities, fetches their sorted sets, and merges them with the regular feed using a two-pointer merge sorted by score descending.

**Trade-off**: slightly more complex read path, but write cost stays O(1) regardless of follower count.

### DB fallback and cache warming

If Redis is cold (key missing) or down:

1. `GetFeed` falls back to a direct PostgreSQL query.
2. On success, a goroutine is launched to write the results back into Redis asynchronously (`WarmCache`), so the next request is served from cache.

The same backfill logic applies to `celebrity_feed`: if the key is missing at read time, `backfillCelebrityFeed` queries the DB and repopulates the sorted set in a background goroutine.

### Cursor pagination

Cursors encode `(created_at, post_id)` as a base64 string. The DB query uses a keyset condition:

```sql
WHERE (created_at < $cursor_time OR (created_at = $cursor_time AND post_id < $cursor_id))
ORDER BY created_at DESC, post_id DESC
```

This avoids `OFFSET` scans and stays stable under concurrent inserts.

---

## Live counters

PostgreSQL `likes_count` and `comments_count` columns are denormalised counters maintained by DB triggers on hard INSERT/DELETE. However, soft-deleted comments (`deleted_at IS NOT NULL`) do not fire the trigger, so the DB column can lag.

Redis is used as the live source of truth for reads:

| Operation | Redis command | Key |
|---|---|---|
| Post liked | `INCR` | `likes_count:{postID}` |
| Post unliked | `DECR` | `likes_count:{postID}` |
| Comment created | `INCR` | `comments_count:{postID}` |
| Comment deleted (soft) | `DECR` | `comments_count:{postID}` |
| Comment liked | `INCR` | `comment_likes_count:{commentID}` |
| Comment unliked | `DECR` | `comment_likes_count:{commentID}` |

On `GET /posts` and `GET /feed`, the service issues a single `MGET` for all post IDs in the page, then overwrites the DB value in memory before serialising. If Redis is unavailable, the DB value is used as fallback (fail-open, see below).

---

## Notifications — async event workers

Every user action that should produce a notification publishes an event to a typed in-process bus (`events.TypedBus[T]`):

| Action | Event type | Bus |
|---|---|---|
| Post liked (by someone else) | `PostLikedEvent` | `likedBus` |
| Comment created | `CommentCreatedEvent` | `commentBus` |
| Comment liked | `CommentLikedEvent` | `commentLikedBus` |
| User followed | `UserFollowedEvent` | `followedBus` |

A single worker goroutine per bus calls the corresponding `NotificationService.Handle*` method, which inserts a row into `public.notifications`.

**Self-notification guard**: services check `authorID != actorID` before publishing the event. The event is never put on the bus, so no row is ever inserted. The DB constraint `chk_notif_not_self` is a safety net, not the primary guard.

**Why async?** Inserting a notification row is not on the critical path of the action that triggered it. Keeping it async means the HTTP response returns immediately without waiting for the notification write.

---

## Rate limiting

Rate limits are enforced per IP using Redis:

```
Key:   ratelimit:{ruleName}:{ip}
Value: INCR on each request
TTL:   set once on first INCR (count == 1), never reset
```

Setting the TTL only on `count == 1` is deliberate: if the TTL were reset on every request, a client spamming the endpoint would never be unblocked. The window slides forward naturally as the key expires.

Current limits:

| Rule | Limit | Window |
|---|---|---|
| `api_global` | 10 req | 1 second |
| `auth_login` | 3 req | 1 minute |
| `auth_register` | 2 req | 1 minute |

If Redis is unavailable, the rate limiter fails open (requests pass through) rather than blocking all traffic.

---

## Auth — JWT + Redis blacklist

JWTs are stateless by design, but logout needs server-side invalidation. The approach:

- On logout, the token's `jti` (or the raw token) is stored in Redis with a TTL equal to the token's remaining lifetime.
- On every protected request, the middleware checks Redis before trusting the token.

**Fail-open on Redis down**: if the blacklist check fails, the request is allowed through. The alternative — blocking all traffic when Redis is down — would be worse. This is a conscious trade-off: a logged-out token could be replayed for a short window during a Redis outage.

---

## Graceful shutdown

All event worker goroutines register themselves with a `sync.WaitGroup` before starting. On SIGINT/SIGTERM:

1. The HTTP server stops accepting new connections (`srv.Shutdown`).
2. The event context is cancelled (`eventsCancel()`), which signals all workers to drain their current event and exit.
3. `eventsWG.Wait()` blocks until every worker has returned.
4. A 10-second timeout prevents hanging forever if a handler is stuck.

This ensures no in-flight notification or fanout write is dropped on clean shutdown.

---

## Database migrations

Migrations are plain numbered SQL files in `db/migrations/`. A lightweight Go runner (`cmd/migrate/main.go`) tracks applied versions in `public.schema_migrations` and wraps each migration in a transaction — the migration either fully applies or fully rolls back.

Three subcommands:

- `up` — apply pending migrations in order
- `stamp` — mark all as applied without running SQL (for baseline of an existing DB)
- `status` — show applied / pending state

No external dependency (no `golang-migrate`, no Atlas). The runner is ~120 lines of standard library Go.

---

## Redis resilience — fail-open

Every Redis call that is not on the write path of a user action is fail-open:

| Component | Redis down behaviour |
|---|---|
| Feed cache | Falls back to DB query, warms cache async |
| Live counters | Falls back to DB column value |
| Rate limiter | Passes all requests through |
| JWT blacklist | Passes all requests through |
| Fanout worker | Logs error, continues processing next event |

The only hard dependency on Redis is the event bus buffer — but that is in-process memory, not Redis.

---

## Known limitations and conscious trade-offs

### In-process event bus — no durability

`events.TypedBus[T]` is a buffered Go channel. Events that sit in the buffer at crash or kill time are lost. There is no replay, no persistence, no at-least-once delivery guarantee.

**Why accepted**: adding Redis Streams or a message broker (NATS, Kafka) would solve this but significantly increases operational complexity. For the current scope — a portfolio project demonstrating the patterns — the trade-off is explicitly accepted. A production system serving real users would need durable event delivery.

**Migration path**: replace `TypedBus.Publish` with a Redis `XADD` call and replace workers with `XREADGROUP` consumers. The handler and service interfaces would not change.

### Denormalised counters can drift

If the process crashes after a Redis `INCR` but before the DB transaction commits (or vice versa), the Redis counter and the DB column can diverge. There is no reconciliation job.

**Why accepted**: the drift is bounded (at most one operation per crash) and Redis is always the read source, so users see consistent values between requests. A periodic sync job could reconcile if needed.

### pushThreshold is a compile-time constant

The celebrity threshold (`pushThreshold = 1000`) is a package-level constant in `feed/fanout.go`. Changing it requires a redeploy.

**Why accepted**: the threshold is not expected to change frequently and making it configurable via environment variable is a straightforward future addition.
