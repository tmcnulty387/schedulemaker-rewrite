package main

/// processDump.php

import (
	"context"
	"database/sql"
	"log"
	"os"

	"rewrite/internal/config"
	"rewrite/internal/tools"
)


func main() {
	ctx := context.Background()

	cfg := config.Load()
	dbConn, err := sql.Open("mysql", cfg.GetDataSourceName())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	p := tools.NewParser(ctx, dbConn, os.Args)

	p.ParseDumps(ctx, &cfg)
	p.ParseDB(ctx)
}
