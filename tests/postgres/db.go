package postgres

import (
	"context"
	"fmt"
	"flash-card-manager/pkg/db"
	"log"
	"strings"
	"testing"
)

type TDB struct {
	DB db.DatabaseInterface
}

func NewFromEnv(dsn string) *TDB {
	db, err := db.NewDB(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to create database connection: %v", err)
	}
	return &TDB{DB: db}
}

func (d *TDB) SetUp(t *testing.T) error {
	t.Helper()
	if d.DB == nil {
		return fmt.Errorf("database is not initialized")
	}
	return nil
}

func (d *TDB) TearDown(ctx context.Context, t *testing.T) error {
	t.Helper()
	if d.DB == nil {
		return fmt.Errorf("database is not initialized")
	}
	return d.Truncate(ctx, t)
}

func (d *TDB) Truncate(ctx context.Context, t *testing.T) error {
	t.Helper()
	if d.DB == nil {
		return fmt.Errorf("database is not initialized")
	}

	var tables []string
	err := d.DB.Select(ctx, &tables, "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'")
	if err != nil {
		return fmt.Errorf("error fetching table names: %w", err)
	}

	if len(tables) == 0 {
		return fmt.Errorf("no tables found: please run migration")
	}

	_, err = d.DB.Exec(ctx, "SET session_replication_role = 'replica';")
	if err != nil {
		return fmt.Errorf("error disabling foreign key check: %w", err)
	}

	q := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(tables, ","))
	if _, err := d.DB.Exec(ctx, q); err != nil {
		return fmt.Errorf("error truncating tables: %w", err)
	}

	_, err = d.DB.Exec(ctx, "SET session_replication_role = 'origin';")
	if err != nil {
		return fmt.Errorf("error enabling foreign key check: %w", err)
	}
	return nil
}
