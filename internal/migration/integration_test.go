package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	migrate "github.com/rubenv/sql-migrate"
	assets "mf-importer/migration"
)

// Only an explicitly supplied disposable test server is used. Each test owns
// a newly created database; no application database is read or modified.
func testDatabase(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MF_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set MF_MIGRATION_TEST_DSN to a disposable MariaDB server")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid test DSN")
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("open test server")
	}
	name := fmt.Sprintf("mf_migration_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		admin.Close()
		t.Fatal("create test database", NewError("test setup", err))
	}
	cfg.DBName = name
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("open test database")
	}
	t.Cleanup(func() {
		db.Close()
		_, err := admin.Exec("DROP DATABASE " + name)
		admin.Close()
		if err != nil {
			t.Error("drop test database")
		}
	})
	return db
}

func assertRun(t *testing.T, db *sql.DB, direction migrate.MigrationDirection, limit, want int) {
	t.Helper()
	n, err := Run(context.Background(), db, direction, limit)
	if err != nil || n != want {
		t.Fatalf("migration count=%d want=%d error=%v", n, want, err)
	}
}

func TestIntegrationLifecycle(t *testing.T) {
	db := testDatabase(t)
	// Legacy library invocation represents the existing CLI's history format.
	source := &migrate.EmbedFileSystemMigrationSource{FileSystem: assets.Files, Root: "db"}
	if n, err := migrate.ExecMax(db, "mysql", source, migrate.Up, 3); err != nil || n != 3 {
		t.Fatal("seed legacy history failed")
	}
	assertRun(t, db, migrate.Up, 0, 4)
	assertRun(t, db, migrate.Up, 0, 0)
	assertRun(t, db, migrate.Down, 1, 1)
	assertRun(t, db, migrate.Up, 0, 1)
	assertRun(t, db, migrate.Down, 0, 7)
	assertRun(t, db, migrate.Up, 0, 7)
}

func TestIntegrationConcurrentStartup(t *testing.T) {
	db := testDatabase(t)
	var wg sync.WaitGroup
	results := make(chan int, 2)
	errs := make(chan error, 2)
	for range 2 {
		wg.Go(func() { n, err := Run(context.Background(), db, migrate.Up, 0); results <- n; errs <- err })
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	total := 0
	for n := range results {
		total += n
	}
	if total != 7 {
		t.Fatalf("applied %d times", total)
	}
}

func TestIntegrationUnknownHistory(t *testing.T) {
	db := testDatabase(t)
	assertRun(t, db, migrate.Up, 0, 7)
	if _, err := db.Exec("INSERT INTO gorp_migrations (id, applied_at) VALUES ('999_dummy.sql', NOW())"); err != nil {
		t.Fatal("seed unknown history")
	}
	n, err := Run(context.Background(), db, migrate.Up, 0)
	if n != 0 || err == nil {
		t.Fatal("unknown history accepted")
	}
}

func TestIntegrationPartialFailureReleasesLock(t *testing.T) {
	db := testDatabase(t)
	source := &migrate.MemoryMigrationSource{Migrations: []*migrate.Migration{{
		Id: "001_dummy.sql", Up: []string{"CREATE TABLE dummy (id INT)", "INVALID SYNTHETIC SQL"},
	}}}
	n, err := run(context.Background(), db, source, migrate.Up, 0, 0)
	if n != 0 || err == nil || !strings.Contains(err.Error(), "001_dummy.sql") {
		t.Fatal("expected migration failure", err)
	}
	var tables int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'dummy'").Scan(&tables); err != nil || tables != 1 {
		t.Fatal("DDL partial application not detected")
	}
	var history int
	if err := db.QueryRow("SELECT COUNT(*) FROM gorp_migrations").Scan(&history); err != nil || history != 0 {
		t.Fatal("failed migration marked applied")
	}
	// A fresh connection can immediately acquire the same lock after failure.
	source.Migrations[0].Up = []string{"ALTER TABLE dummy ADD COLUMN note TEXT"}
	n, err = run(context.Background(), db, source, migrate.Up, 0, 0)
	if n != 1 || err != nil {
		t.Fatal("lock leaked after failure", err)
	}
}

func TestIntegrationLockTimeout(t *testing.T) {
	db := testDatabase(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan error, 1)
	source := blockingSource{entered: entered, release: release}
	go func() { _, err := run(context.Background(), db, source, migrate.Up, 0, 0); finished <- err }()
	select {
	case <-entered:
	case err := <-finished:
		t.Fatal("lock holder failed", err)
	case <-time.After(10 * time.Second):
		t.Fatal("lock holder did not start")
	}
	n, err := run(context.Background(), db, &migrate.MemoryMigrationSource{}, migrate.Up, 0, 0)
	close(release)
	holderErr := <-finished
	if holderErr != nil {
		t.Fatal(holderErr)
	}
	if n != 0 || err == nil || !strings.Contains(err.Error(), "lock acquisition") {
		t.Fatal("contending migration did not stop", err)
	}
}

type blockingSource struct {
	entered chan struct{}
	release chan struct{}
}

func (s blockingSource) FindMigrations() ([]*migrate.Migration, error) {
	close(s.entered)
	<-s.release
	return nil, nil
}

func TestIntegrationLostSessionCannotContinue(t *testing.T) {
	db := testDatabase(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan error, 1)
	source := blockingSource{entered: entered, release: release}
	go func() { _, err := run(context.Background(), db, source, migrate.Up, 0, 0); finished <- err }()
	select {
	case <-entered:
	case err := <-finished:
		t.Fatal("lock holder failed", err)
	case <-time.After(10 * time.Second):
		t.Fatal("lock holder did not start")
	}
	var name string
	if err := db.QueryRow("SELECT DATABASE()").Scan(&name); err != nil {
		close(release)
		t.Fatal("read test database name")
	}
	digest := sha256.Sum256([]byte(name))
	lockName := fmt.Sprintf("mf-importer:migrate:%x", digest[:16])
	var connectionID int64
	if err := db.QueryRow("SELECT IS_USED_LOCK(?)", lockName).Scan(&connectionID); err != nil {
		close(release)
		t.Fatal("find test lock owner")
	}
	_, killErr := db.Exec(fmt.Sprintf("KILL CONNECTION %d", connectionID))
	close(release)
	err := <-finished
	if killErr != nil {
		t.Fatal("kill test connection failed")
	}
	if err == nil {
		t.Fatal("continued after losing the locked session")
	}
	// The next run must obtain a fresh lock and still be able to apply all SQL.
	assertRun(t, db, migrate.Up, 0, 7)
}
