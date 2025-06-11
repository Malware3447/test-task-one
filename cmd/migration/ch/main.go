package main

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"log"
	"os"
	"strings"
)

func main() {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "demo",
			Password: "demo",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Read migration file
	migration, err := os.ReadFile("migrations/ch/init.sql")
	if err != nil {
		log.Fatal(err)
	}

	queries := strings.Split(string(migration), ";")

	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		if err := conn.Exec(ctx, query); err != nil {
			log.Fatalf("Failed to execute query: %s, error: %v", query, err)
		}
	}
	fmt.Println("Миграция успешно выполнена")
}
