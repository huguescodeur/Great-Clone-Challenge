# insta-like

REST API for an Instagram-like social platform. Built with Go, Chi, and PostgreSQL.

## Features

- Authentication with JWT (register / login)
- Posts with media attachments and cursor-based pagination
- Reactions on posts (like, love, laugh)
- Nested comments with replies and cursor-based pagination
- Reactions on comments
- Follow / unfollow system with status tracking

## Tech stack

- **Language**: Go
- **Router**: go-chi/chi
- **Database**: PostgreSQL (pgx driver)
- **Auth**: JWT (golang-jwt/jwt)
- **Validation**: go-playground/validator
- **Docs**: Swagger (swaggo/swag)

## Prerequisites

- Go 1.21+
- PostgreSQL running locally

## Getting started

1. Create a PostgreSQL database named `insta`.

2. Load the schema:
   ```
   psql -h localhost -U postgres -d insta -f db/schema.sql
   ```

3. Update the connection string in `cmd/server/main.go` if needed:
   ```
   postgres://postgres:<password>@localhost:5432/insta
   ```

4. Run the server:
   ```
   go run ./cmd/server
   ```

The server starts on port `8080` by default. Set the `PORT` environment variable to change it.

## API documentation

Swagger UI is available at:
```
http://localhost:8080/swagger/index.html
```

To regenerate the docs after modifying annotations:
```
swag init -g cmd/server/main.go --output docs
```

## Project structure

```
cmd/server/        Entry point
db/                SQL schema
internal/
  app/             App wiring and routes
  auth/            Register, login, JWT middleware
  user/            User model and store
  post/            Post CRUD
  postmedia/       Media attachments for posts
  like/            Reactions on posts
  comment/         Comments and replies
  commentlike/     Reactions on comments
  follow/          Follow / unfollow
  middlewares/     Logger, rate limiter
  pkg/             Shared utilities (errors, context keys, responses)
docs/              Generated Swagger files
```

## Database schema

The schema file is at `db/schema.sql`. It requires the `citext` extension (included in the file).

### users

| Column | Type | Notes |
|--------|------|-------|
| user_id | uuid | Primary key |
| username | citext | Unique, case-insensitive |
| email | citext | Unique, case-insensitive |
| password_hash | varchar(255) | Bcrypt hash |
| full_name | varchar(255) | |
| bio | text | Max 500 chars |
| profile_pic_url | varchar(500) | |
| followers_count | bigint | Denormalized counter |
| following_count | bigint | Denormalized counter |
| posts_count | bigint | Denormalized counter |
| created_at | timestamp | |
| updated_at | timestamp | |

### posts

| Column | Type | Notes |
|--------|------|-------|
| post_id | uuid | Primary key |
| user_id | uuid | FK -> users |
| content | text | Max 5000 chars |
| likes_count | bigint | Denormalized counter |
| comments_count | bigint | Denormalized counter |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | Soft delete |

### post_media

| Column | Type | Notes |
|--------|------|-------|
| media_id | uuid | Primary key |
| post_id | uuid | FK -> posts (cascade) |
| url | text | |
| type | text | `image` or `video` |
| created_at | timestamptz | |

### likes

| Column | Type | Notes |
|--------|------|-------|
| post_id | uuid | PK + FK -> posts (cascade) |
| user_id | uuid | PK + FK -> users (cascade) |
| reaction_type | varchar(20) | `like`, `love`, `laugh` |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### comments

| Column | Type | Notes |
|--------|------|-------|
| comment_id | uuid | Primary key |
| post_id | uuid | FK -> posts (cascade) |
| user_id | uuid | FK -> users |
| parent_comment_id | uuid | FK -> comments (nullable, for replies) |
| content | text | Max 2000 chars |
| likes_count | bigint | Denormalized counter |
| created_at | timestamp | |
| updated_at | timestamp | |
| deleted_at | timestamp | Soft delete |

### comment_likes

| Column | Type | Notes |
|--------|------|-------|
| comment_id | uuid | PK + FK -> comments (cascade) |
| user_id | uuid | PK + FK -> users |
| reaction_type | varchar(20) | `like`, `love`, `laugh` |
| created_at | timestamp | |
| updated_at | timestamp | |

### followers

| Column | Type | Notes |
|--------|------|-------|
| follower_id | uuid | PK + FK -> users (cascade) |
| followee_id | uuid | PK + FK -> users (cascade) |
| status | follow_status | `pending` or `accepted` |
| created_at | timestamp | |

Self-follow is prevented by a CHECK constraint. The composite primary key `(follower_id, followee_id)` ensures uniqueness.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /api/v1/auth/register | No | Register a new user |
| POST | /api/v1/auth/login | No | Login and get JWT |
| GET | /api/v1/posts | Yes | Paginated feed |
| GET | /api/v1/posts/{id} | Yes | Posts by user |
| POST | /api/v1/posts | Yes | Create a post |
| PATCH | /api/v1/posts | Yes | Update a post |
| DELETE | /api/v1/posts | Yes | Delete a post |
| GET | /api/v1/posts/{postID}/likes | Yes | Likes on a post |
| GET | /api/v1/posts/{postID}/likes/me | Yes | My reaction on a post |
| POST | /api/v1/posts/{postID}/likes | Yes | React to a post |
| PATCH | /api/v1/posts/{postID}/likes | Yes | Update reaction |
| DELETE | /api/v1/posts/{postID}/likes | Yes | Remove reaction |
| GET | /api/v1/posts/{postID}/comments | Yes | Comments on a post |
| POST | /api/v1/posts/{postID}/comments | Yes | Add a comment or reply |
| GET | /api/v1/comments/{id}/replies | Yes | Replies to a comment |
| PATCH | /api/v1/comments/{id} | Yes | Update a comment |
| DELETE | /api/v1/comments/{id} | Yes | Delete a comment |
| GET | /api/v1/comments/{commentID}/likes | Yes | Likes on a comment |
| GET | /api/v1/comments/{commentID}/likes/me | Yes | My reaction on a comment |
| POST | /api/v1/comments/{commentID}/likes | Yes | React to a comment |
| PATCH | /api/v1/comments/{commentID}/likes | Yes | Update reaction |
| DELETE | /api/v1/comments/{commentID}/likes | Yes | Remove reaction |
| POST | /api/v1/users/{userID}/follow | Yes | Follow a user |
| DELETE | /api/v1/users/{userID}/follow | Yes | Unfollow a user |
| GET | /api/v1/users/{userID}/follow/status | Yes | Check follow status |
| GET | /api/v1/users/{userID}/followers | Yes | List followers |
| GET | /api/v1/users/{userID}/following | Yes | List following |

Protected routes require the `Authorization: Bearer <token>` header.
