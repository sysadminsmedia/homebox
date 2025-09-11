package main

import (
	"context"
	"log"
	"os"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	_ "github.com/sysadminsmedia/homebox/backend/internal/data/ent/migrate"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("postgres", "host=localhost port=5432 user=homebox dbname=homebox password=homebox sslmode=disable")
	if err != nil {
		log.Fatalf("failed connecting to mysql: %v", err)
	}
	defer client.Close()
	ctx := context.Background()
	// Dump migration changes to an SQL script.
	f, err := os.Create("migrate.sql")
	if err != nil {
		log.Fatalf("create migrate file: %v", err)
	}
	defer f.Close()
	if err := client.Schema.WriteTo(ctx, f); err != nil {
		log.Fatalf("failed printing schema changes: %v", err)
	}
}
