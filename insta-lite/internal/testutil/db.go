package testutil

// ? go:build integration

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func TestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open test DB: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("test DB unavailable (%v) — skipping integration test", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{Addr: addr})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("test Redis unavailable (%v) — skipping integration test", err)
	}

	t.Cleanup(func() { client.Close() })
	return client
}

func InsertTestUser(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO users (user_id, username, email, password_hash, full_name)
		VALUES ($1, $2, $3, 'testhash', 'Test User')
	`, id, "tu_"+id.String()[:8], "tu_"+id.String()[:8]+"@test.local")
	if err != nil {
		t.Fatalf("InsertTestUser: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM users WHERE user_id = $1", id)
	})
	return id
}

func InsertTestPost(t *testing.T, db *sql.DB, userID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO posts (post_id, user_id, content) VALUES ($1, $2, 'integration test post')
	`, id, userID)
	if err != nil {
		t.Fatalf("InsertTestPost: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM posts WHERE post_id = $1", id)
	})
	return id
}

func InsertTestComment(t *testing.T, db *sql.DB, postID, userID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO comments (comment_id, post_id, user_id, content)
		VALUES ($1, $2, $3, 'integration test comment')
	`, id, postID, userID)
	if err != nil {
		t.Fatalf("InsertTestComment: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM comments WHERE comment_id = $1", id)
	})
	return id
}

func CleanRedisKey(t *testing.T, client *redis.Client, key string) {
	t.Helper()
	t.Cleanup(func() {
		client.Del(context.Background(), key)
	})
}
