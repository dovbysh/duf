package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func (p *PostgresClient) Migrate() error {
	m, err := p.newMigrator()
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func (p *PostgresClient) MigrateDownOne() error {
	m, err := p.newMigrator()
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback one migration: %w", err)
	}

	return nil
}

func (p *PostgresClient) ApplyDirtyMigration() (bool, error) {
	m, err := p.newMigrator()
	if err != nil {
		return false, err
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		return false, fmt.Errorf("read migration version: %w", err)
	}
	if !dirty {
		return false, nil
	}

	previousVersion := int(version) - 1
	if version == 0 {
		previousVersion = -1
	}

	if err := m.Force(previousVersion); err != nil {
		return false, fmt.Errorf("clear dirty migration version %d: %w", version, err)
	}
	if err := m.Steps(1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return false, fmt.Errorf("apply dirty migration %d: %w", version, err)
	}

	return true, nil
}

func (p *PostgresClient) newMigrator() (*migrate.Migrate, error) {
	migrationDB, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return nil, fmt.Errorf("open migration database: %w", err)
	}
	migrationDB.SetMaxOpenConns(1)
	migrationDB.SetMaxIdleConns(1)

	driver, err := postgres.WithInstance(migrationDB, &postgres.Config{
		MigrationsTable:       `"public"."duf_migrations"`,
		MigrationsTableQuoted: true,
	})
	if err != nil {
		migrationDB.Close()
		return nil, fmt.Errorf("create postgres migration driver: %w", err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		driver.Close()
		return nil, fmt.Errorf("create embedded migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		source.Close()
		driver.Close()
		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return m, nil
}
