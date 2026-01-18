package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"phsio_track_backend/internal/config"
	"phsio_track_backend/internal/repo"
)

// bootstrap creates tables/indexes and exits.
func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := repo.NewDB(ctx, repo.DBConfig{
		User:          cfg.DBUser,
		Password:      cfg.DBPassword,
		ConnectString: cfg.DBConnectString,
		TNSAdmin:      cfg.TNSAdmin,
	})
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer db.Close()

	if err := repo.BootstrapSchema(ctx, db); err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	fmt.Println("bootstrap complete")
}
