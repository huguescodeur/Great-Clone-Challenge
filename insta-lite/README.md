# insta-lite

REST API for an Instagram-like social platform. Built with Go, Chi, PostgreSQL, and Redis.

## Features

- JWT authentication (register / login / logout with token blacklist)
- Posts with media attachments and cursor-based pagination
- Reactions on posts and comments
- Nested comments with replies and cursor-based pagination
- Follow / unfollow system
- Personalised feed with hybrid push/pull fanout (Redis sorted sets)
- In-memory live counters for likes and comments (Redis INCR/DECR)
- Notifications (like, comment, follow, comment like) via async event workers
- Rate limiting per IP stored in Redis
- Graceful shutdown with WaitGroup across all event workers
- Database migrations with a built-in Go runner

## Tech stack

| Layer | Choice |
|---|---|
| Language | Go 1.26 |
| Router | go-chi/chi v5 |
| Database | PostgreSQL 18 (pgx v5 driver) |
| Cache / counters | Redis 7 |
| Auth | golang-jwt/jwt v5 + bcrypt |
| Validation | go-playground/validator v10 |
| Docs | Swagger (swaggo/swag) |
| Containerisation | Docker + docker-compose |

## Prerequisites

- Docker and docker-compose (recommended)
- Go 1.26+ (only needed for local dev outside Docker)

## Getting started

### With Docker (recommended)

```bash
cp .env.example .env          # fill in DB_PASSWORD and JWT_SECRET
make up                        # starts postgres, redis, api
make migrate-docker-stamp      # baseline the DB if it already has the schema
# OR on a fresh DB:
make migrate-docker            # applies all migrations
```

The API is available at `http://localhost:8080/api/v1`.

### Local dev (without Docker)

```bash
cp .env.example .env           # set DATABASE_URL and REDIS_ADDR
make run-db                    # start postgres and redis via Docker
make migrate                   # apply migrations against local DB
make run-api                   # start the Go server
```

## Makefile targets

| Target | Description |
|---|---|
| `make up` | Start all containers in background |
| `make down` | Stop containers |
| `make logs` | Follow API logs |
| `make ps` | Container status |
| `make run-db` | Start only postgres and redis |
| `make run-api` | Run API locally (loads `.env`) |
| `make build` | Build binary to `bin/server` |
| `make migrate` | Apply pending migrations (local DB) |
| `make migrate-docker` | Apply pending migrations (Docker DB on port 5433) |
| `make migrate-stamp` | Baseline existing local DB |
| `make migrate-docker-stamp` | Baseline existing Docker DB |
| `make migrate-status` | Show applied / pending migrations |
| `make swag` | Regenerate Swagger docs |
| `make clean` | Remove containers, volumes, and `bin/` |

## Environment variables

| Variable | Description | Default |
|---|---|---|
| `DATABASE_URL` | PostgreSQL DSN | required |
| `REDIS_ADDR` | Redis address (`host:port`) | `localhost:6379` |
| `PORT` | HTTP port | `8080` |
| `JWT_SECRET` | JWT signing secret | required |
| `DB_PASSWORD` | Used by docker-compose for postgres | required |

## Database migrations

Migrations live in `db/migrations/` as numbered SQL files. The runner is at `cmd/migrate/main.go`.

```
db/migrations/
  000001_init_schema.sql       — extensions, types, tables, indexes, triggers
  000002_add_notifications.sql — notification_type enum, notifications table
```

Applied versions are tracked in `public.schema_migrations`.

### Commands

```bash
make migrate          # apply pending migrations
make migrate-status   # list applied / pending
make migrate-stamp    # mark all as applied without running SQL (baseline)
```

## API documentation

Swagger UI is served at:

```
http://localhost:8080/swagger/index.html
```

To regenerate after modifying handler annotations:

```bash
make swag
```

## Project structure

```
cmd/
  server/          Entry point (HTTP server, graceful shutdown)
  migrate/         Standalone migration runner
db/
  schema.sql       Full schema snapshot (source of truth for reference)
  migrations/      Versioned SQL migration files
internal/
  app/             Dependency wiring and route registration
  auth/            Register, login, JWT middleware, Redis token blacklist
  user/            User model and store
  post/            Post CRUD and cursor pagination
  postmedia/       Media attachments
  like/            Reactions on posts (with Redis live counter)
  comment/         Comments and replies (with Redis live counter)
  commentlike/     Reactions on comments (with Redis live counter)
  follow/          Follow / unfollow
  feed/            Personalised feed — fanout service, Redis cache, hybrid push/pull
  notification/    Notification model, store, service, handler
  events/          Typed async event buses and generic worker pool
  middlewares/     Request logger, configurable Redis rate limiter
  pkg/             Shared utilities (errors, context keys, response helpers)
docs/              Generated Swagger files
```

## Feed architecture

The feed uses a **hybrid fanout** strategy:

- **Push (write fanout)**: when a regular user posts, the post ID is pushed into each follower's sorted set `feed:{followerID}` in Redis (score = UnixNano).
- **Pull (read fanout)**: when a user with more than `pushThreshold` followers posts, the post ID is written once into `celebrity_feed:{authorID}`. On read, the service merges the user's regular feed with the celebrity feeds of followed accounts.
- **DB fallback**: if Redis is down or the cache is cold, the service falls back to a direct DB query and warms the cache asynchronously in a goroutine.

## Live counters

`likes_count`, `comments_count`, and `comment_likes_count` are maintained in Redis using `INCR` / `DECR` and read back with `MGET` in batch. The values shown in `GET /posts` and `GET /feed` always reflect Redis, not the potentially stale DB column. PostgreSQL triggers keep the DB columns consistent for hard-delete operations.

## Notifications

Notifications are published asynchronously via in-process typed event buses (`events.TypedBus[T]`). Each action (post like, comment, comment like, follow) publishes an event; dedicated worker goroutines consume the bus and insert rows into `public.notifications`. Workers participate in the WaitGroup so graceful shutdown drains the queue before exit. Self-notifications are suppressed at the service layer before the event is even published.

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | No | Create account |
| POST | `/auth/login` | No | Login, returns JWT |
| GET | `/feed` | Yes | Personalised paginated feed |
| GET | `/posts` | Yes | All posts (paginated) |
| GET | `/posts/{id}` | Yes | Posts by user ID |
| POST | `/posts` | Yes | Create a post |
| PATCH | `/posts/{id}` | Yes | Update a post |
| DELETE | `/posts/{id}` | Yes | Delete a post |
| GET | `/posts/{postID}/likes` | Yes | Likes on a post |
| GET | `/posts/{postID}/likes/me` | Yes | My reaction on a post |
| POST | `/posts/{postID}/likes` | Yes | React to a post |
| PATCH | `/posts/{postID}/likes` | Yes | Update reaction |
| DELETE | `/posts/{postID}/likes` | Yes | Remove reaction |
| GET | `/posts/{postID}/comments` | Yes | Comments on a post |
| POST | `/posts/{postID}/comments` | Yes | Add a comment or reply |
| GET | `/comments/{id}/replies` | Yes | Replies to a comment |
| PATCH | `/comments/{id}` | Yes | Update a comment |
| DELETE | `/comments/{id}` | Yes | Soft-delete a comment |
| GET | `/comments/{commentID}/likes` | Yes | Likes on a comment |
| GET | `/comments/{commentID}/likes/me` | Yes | My reaction on a comment |
| POST | `/comments/{commentID}/likes` | Yes | React to a comment |
| PATCH | `/comments/{commentID}/likes` | Yes | Update reaction |
| DELETE | `/comments/{commentID}/likes` | Yes | Remove reaction |
| POST | `/users/{userID}/follow` | Yes | Follow a user |
| DELETE | `/users/{userID}/follow` | Yes | Unfollow a user |
| GET | `/users/{userID}/follow/status` | Yes | Check follow status |
| GET | `/users/{userID}/followers` | Yes | List followers |
| GET | `/users/{userID}/following` | Yes | List following |
| GET | `/notifications` | Yes | List notifications (paginated) |
| PATCH | `/notifications/{id}/read` | Yes | Mark notification as read |

Protected routes require `Authorization: Bearer <token>`.

## Known limitations

The event buses (`events.TypedBus[T]`) are **in-process buffered channels**. Events that are in the buffer at crash time are lost — there is no replay or persistence guarantee. Migrating to Redis Streams or a message broker (NATS, Kafka) would make the pipeline durable but adds operational overhead that is out of scope for this project.
