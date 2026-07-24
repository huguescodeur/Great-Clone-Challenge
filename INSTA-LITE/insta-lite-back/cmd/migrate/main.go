package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationsDir = "db/migrations"

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if err := ensureMigrationsTable(db); err != nil {
		log.Fatalf("create schema_migrations: %v", err)
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		runUp(db)
	case "stamp":
		runStamp(db)
	case "status":
		runStatus(db)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\nusage: migrate [up|stamp|status]\n", cmd)
		os.Exit(1)
	}
}

func runUp(db *sql.DB) {
	files, err := sqlFiles()
	if err != nil {
		log.Fatalf("list migrations: %v", err)
	}

	applied, skipped := 0, 0
	for _, filename := range files {
		done, err := isApplied(db, filename)
		if err != nil {
			log.Fatalf("check %s: %v", filename, err)
		}
		if done {
			fmt.Printf("  skip   %s\n", filename)
			skipped++
			continue
		}
		if err := applyMigration(db, filename); err != nil {
			log.Fatalf("apply %s: %v", filename, err)
		}
		fmt.Printf("  apply  %s\n", filename)
		applied++
	}

	fmt.Printf("\n%d applied, %d skipped\n", applied, skipped)
}

func runStamp(db *sql.DB) {
	files, err := sqlFiles()
	if err != nil {
		log.Fatalf("list migrations: %v", err)
	}

	stamped, skipped := 0, 0
	for _, filename := range files {
		done, err := isApplied(db, filename)
		if err != nil {
			log.Fatalf("check %s: %v", filename, err)
		}
		if done {
			fmt.Printf("  skip   %s (already stamped)\n", filename)
			skipped++
			continue
		}
		if _, err := db.Exec(
			"INSERT INTO public.schema_migrations (version) VALUES ($1)", filename,
		); err != nil {
			log.Fatalf("stamp %s: %v", filename, err)
		}
		fmt.Printf("  stamp  %s\n", filename)
		stamped++
	}

	fmt.Printf("\n%d stamped, %d skipped\n", stamped, skipped)
}

func runStatus(db *sql.DB) {
	files, err := sqlFiles()
	if err != nil {
		log.Fatalf("list migrations: %v", err)
	}

	fmt.Printf("%-50s  %s\n", "version", "status")
	fmt.Printf("%-50s  %s\n", "-------", "------")
	for _, filename := range files {
		done, err := isApplied(db, filename)
		if err != nil {
			log.Fatalf("check %s: %v", filename, err)
		}
		status := "pending"
		if done {
			status = "applied"
		}
		fmt.Printf("%-50s  %s\n", filename, status)
	}
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS public.schema_migrations (
			version     text        NOT NULL,
			applied_at  timestamptz NOT NULL DEFAULT now(),
			CONSTRAINT schema_migrations_pkey PRIMARY KEY (version)
		)
	`)
	return err
}

func sqlFiles() ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func isApplied(db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM public.schema_migrations WHERE version = $1)",
		version,
	).Scan(&exists)
	return exists, err
}

func applyMigration(db *sql.DB, filename string) error {
	content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("execute SQL: %w", err)
	}

	if _, err := tx.Exec(
		"INSERT INTO public.schema_migrations (version) VALUES ($1)", filename,
	); err != nil {
		return fmt.Errorf("record version: %w", err)
	}

	return tx.Commit()
}
