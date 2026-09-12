package main

import (
	"context"
	"mf-importer/internal/migration"
	"mf-importer/internal/repository"
	"time"

	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
)

type databaseConfig struct{ host, port, user, pass, name string }

func importerDatabaseConfig() databaseConfig {
	value := func(primary, legacy, fallback string) string {
		if v := envOr(primary, legacy); v != "" {
			return v
		}
		return fallback
	}
	return databaseConfig{
		host: value("DB_HOST", "db_host", "127.0.0.1"),
		port: value("DB_PORT", "db_port", "3306"),
		user: value("DB_USER", "db_user", "root"),
		pass: value("DB_PASS", "db_pass", "password"),
		name: value("DB_NAME", "db_name", "mfimporter"),
	}
}

func openImporterDatabase() (*repository.DBClient, error) {
	c := importerDatabaseConfig()
	db, err := repository.NewDBRepository(c.host, c.port, c.user, c.pass, c.name)
	if err != nil {
		return nil, migration.NewError("database connection", err)
	}
	return db, nil
}

func applyMigrations(ctx context.Context, l *zap.Logger, db *repository.DBClient, direction migrate.MigrationDirection, limit int) error {
	sqlDB, err := db.Conn.DB()
	if err != nil {
		return migration.NewError("database handle", err)
	}
	started := time.Now()
	count, err := migration.Run(ctx, sqlDB, direction, limit)
	fields := []zap.Field{zap.Int("applied", count), zap.Duration("duration", time.Since(started))}
	if err != nil {
		l.Error("DB migration failed", append(fields, zap.Error(err))...)
		return err
	}
	l.Info("DB migration complete", fields...)
	return nil
}
