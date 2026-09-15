package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	migrate "github.com/rubenv/sql-migrate"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"mf-importer/internal/model"
	"mf-importer/internal/repository"
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
	assertRun(t, db, migrate.Up, 0, 6)
	assertRun(t, db, migrate.Up, 0, 0)
	assertRun(t, db, migrate.Down, 1, 1)
	assertRun(t, db, migrate.Up, 0, 1)
	assertRun(t, db, migrate.Down, 0, 9)
	assertRun(t, db, migrate.Up, 0, 9)
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
	if total != 9 {
		t.Fatalf("applied %d times", total)
	}
}

func TestIntegrationUnknownHistory(t *testing.T) {
	db := testDatabase(t)
	assertRun(t, db, migrate.Up, 0, 9)
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
	assertRun(t, db, migrate.Up, 0, 9)
}

func TestIntegrationSbiSchemaUpgradePreservesHistoryAndAddsUniqueness(t *testing.T) {
	db := testDatabase(t)
	source := &migrate.EmbedFileSystemMigrationSource{FileSystem: assets.Files, Root: "db"}
	if n, err := migrate.ExecMax(db, "mysql", source, migrate.Up, 7); err != nil || n != 7 {
		t.Fatal("seed legacy SBI schema failed")
	}
	snapshotID := insertLegacySbiSnapshot(t, db)
	insertLegacySbiHolding(t, db, snapshotID)
	assertRun(t, db, migrate.Up, 0, 2)

	var schemaVersion string
	if err := db.QueryRow("SELECT schema_version FROM sbi_snapshot WHERE id = ?", snapshotID).Scan(&schemaVersion); err != nil {
		t.Fatal("read preserved schema version")
	}
	if schemaVersion != "1" {
		t.Fatalf("schema_version = %q, want preserved integer text", schemaVersion)
	}
	var compositeFIGI *string
	if err := db.QueryRow("SELECT composite_figi FROM sbi_holding WHERE snapshot_id = ?", snapshotID).Scan(&compositeFIGI); err != nil {
		t.Fatal("read preserved holding")
	}
	if compositeFIGI != nil {
		t.Fatalf("legacy composite_figi = %v, want NULL", *compositeFIGI)
	}

	if _, err := db.Exec("UPDATE sbi_holding SET composite_figi = ? WHERE snapshot_id = ?", "DUMMY0000001", snapshotID); err != nil {
		t.Fatal("set synthetic FIGI")
	}
	if _, err := db.Exec("INSERT INTO sbi_holding (snapshot_id, section, composite_figi, name, quantity, unit_cost, unit_price, prev_day_jpy, prev_day_pct, pnl_jpy, pnl_pct, value_jpy) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", snapshotID, "nisa_domestic", "DUMMY0000001", "ダミー銘柄B", 1, 1, 1, 0, 0, 0, 0, 1); err == nil {
		t.Fatal("duplicate snapshot/section/FIGI was accepted")
	}
}

func TestIntegrationSbiJSONRoundTrip(t *testing.T) {
	db := testDatabase(t)
	assertRun(t, db, migrate.Up, 0, 9)
	raw, err := os.ReadFile("../../test/sbi_example_new.json")
	if err != nil {
		t.Fatal("read synthetic fixture")
	}
	snapshot, holdings, err := model.ParseSbiJSON(raw)
	if err != nil {
		t.Fatal("parse synthetic fixture")
	}
	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal("open GORM test client")
	}
	client := &repository.DBClient{Conn: gormDB}
	inserted, err := client.ImportSbiSnapshot(context.Background(), snapshot, holdings)
	if err != nil || !inserted {
		t.Fatalf("round trip import inserted=%v error=%v", inserted, err)
	}
	var gotVersion string
	if err := db.QueryRow("SELECT schema_version FROM sbi_snapshot WHERE id = ?", snapshot.ID).Scan(&gotVersion); err != nil {
		t.Fatal("read round trip snapshot")
	}
	if gotVersion != model.CurrentSbiSchemaVersion {
		t.Fatalf("round trip schema_version = %q", gotVersion)
	}
	rows, err := db.Query("SELECT section, composite_figi FROM sbi_holding WHERE snapshot_id = ?", snapshot.ID)
	if err != nil {
		t.Fatal("query round trip holdings")
	}
	gotFIGIBySection := make(map[string]string)
	for rows.Next() {
		var section, figi string
		if err := rows.Scan(&section, &figi); err != nil {
			t.Fatal("scan round trip holding")
		}
		gotFIGIBySection[section] = figi
	}
	if err := rows.Err(); err != nil {
		t.Fatal("iterate round trip holdings")
	}
	if err := rows.Close(); err != nil {
		t.Fatal("close round trip holdings")
	}
	wantFIGIBySection := map[string]string{
		"nisa_domestic":  "DUMMY0000001",
		"nisa_us":        "DUMMY0000002",
		"nisa_funds":     "DUMMY0000003",
		"old_nisa_funds": "DUMMY0000004",
	}
	if !reflect.DeepEqual(gotFIGIBySection, wantFIGIBySection) {
		t.Fatalf("round trip FIGIs = %#v, want %#v", gotFIGIBySection, wantFIGIBySection)
	}

	snapshotAgain, holdingsAgain, err := model.ParseSbiJSON(raw)
	if err != nil {
		t.Fatal("parse synthetic fixture for duplicate import")
	}
	inserted, err = client.ImportSbiSnapshot(context.Background(), snapshotAgain, holdingsAgain)
	if err != nil || inserted {
		t.Fatalf("duplicate round trip import inserted=%v error=%v", inserted, err)
	}

	snapshotNext, holdingsNext, err := model.ParseSbiJSON(raw)
	if err != nil {
		t.Fatal("parse synthetic fixture for next snapshot")
	}
	snapshotNext.FetchedAt = snapshotNext.FetchedAt.Add(time.Hour)
	*holdingsNext[1].CompositeFIGI = *holdingsNext[0].CompositeFIGI
	inserted, err = client.ImportSbiSnapshot(context.Background(), snapshotNext, holdingsNext)
	if err != nil || !inserted {
		t.Fatalf("next snapshot import inserted=%v error=%v", inserted, err)
	}
	var sameFIGICount int
	if err := db.QueryRow("SELECT COUNT(*) FROM sbi_holding WHERE snapshot_id = ? AND composite_figi = ?", snapshotNext.ID, "DUMMY0000001").Scan(&sameFIGICount); err != nil {
		t.Fatal("count same FIGI across sections")
	}
	if sameFIGICount != 2 {
		t.Fatalf("same FIGI count across sections = %d, want 2", sameFIGICount)
	}
}

func TestIntegrationSbiMigrationDownGuard(t *testing.T) {
	for _, test := range []struct {
		name    string
		sqlMode string
		figi    bool
	}{
		{name: "strict version", sqlMode: "STRICT_ALL_TABLES"},
		{name: "non-strict version", sqlMode: ""},
		{name: "non-null FIGI", sqlMode: "STRICT_ALL_TABLES", figi: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := testDatabase(t)
			db.SetMaxOpenConns(1)
			db.SetMaxIdleConns(1)
			assertRun(t, db, migrate.Up, 8, 8)
			if _, err := db.Exec("SET SESSION sql_mode = ?", test.sqlMode); err != nil {
				t.Fatal("set SQL mode")
			}
			if test.figi {
				snapshotID := insertLegacySbiSnapshot(t, db)
				insertLegacySbiHoldingWithFIGI(t, db, snapshotID, "DUMMY0000001")
			} else if _, err := db.Exec("INSERT INTO sbi_snapshot (fetched_at, status, schema_version) VALUES (?, ?, ?)", "2026-08-16 12:00:00", "ERROR", "not-a-number"); err != nil {
				t.Fatal("seed invalid schema version")
			}
			if _, err := Run(context.Background(), db, migrate.Down, 1); err == nil {
				t.Fatal("unsafe down migration was accepted")
			}
			var columnType string
			if err := db.QueryRow("SELECT data_type FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sbi_snapshot' AND column_name = 'schema_version'").Scan(&columnType); err != nil {
				t.Fatal("read schema version type")
			}
			if columnType != "varchar" {
				t.Fatalf("schema_version type = %q after refused down", columnType)
			}
			var figiColumnCount int
			if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sbi_holding' AND column_name = 'composite_figi'").Scan(&figiColumnCount); err != nil {
				t.Fatal("read FIGI column")
			}
			if figiColumnCount != 1 {
				t.Fatalf("FIGI column count = %d after refused down", figiColumnCount)
			}
			var indexCount int
			if err := db.QueryRow("SELECT COUNT(DISTINCT index_name) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'sbi_holding' AND index_name IN ('uq_snapshot_section_figi', 'idx_composite_figi')").Scan(&indexCount); err != nil {
				t.Fatal("read FIGI indexes")
			}
			if indexCount != 2 {
				t.Fatalf("FIGI index count = %d after refused down", indexCount)
			}
			var migrationCount int
			if err := db.QueryRow("SELECT COUNT(*) FROM gorp_migrations WHERE id = ?", "008_sbi_schema_version_figi.sql").Scan(&migrationCount); err != nil {
				t.Fatal("read migration history")
			}
			if migrationCount != 1 {
				t.Fatalf("migration history count = %d after refused down, want 1", migrationCount)
			}
			if test.figi {
				var gotFIGI string
				if err := db.QueryRow("SELECT composite_figi FROM sbi_holding LIMIT 1").Scan(&gotFIGI); err != nil {
					t.Fatal("read preserved FIGI")
				}
				if gotFIGI != "DUMMY0000001" {
					t.Fatalf("preserved FIGI = %q", gotFIGI)
				}
			}
		})
	}
}

func insertLegacySbiSnapshot(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	columns := []string{"fetched_at", "status", "schema_version", "grand_total_jpy", "nisa_total_jpy", "nisa_prev_day_jpy", "nisa_prev_day_pct", "nisa_prev_month_jpy", "nisa_prev_month_pct", "nisa_pnl_jpy", "nisa_pnl_pct", "nisa_domestic_value_jpy", "nisa_domestic_pnl_jpy", "nisa_domestic_pnl_pct", "nisa_domestic_prev_day_jpy", "nisa_domestic_prev_day_pct", "nisa_domestic_prev_month_jpy", "nisa_domestic_prev_month_pct", "nisa_us_value_jpy", "nisa_us_pnl_jpy", "nisa_us_pnl_pct", "nisa_us_prev_day_jpy", "nisa_us_prev_day_pct", "nisa_us_prev_month_jpy", "nisa_us_prev_month_pct", "nisa_funds_value_jpy", "nisa_funds_pnl_jpy", "nisa_funds_pnl_pct", "nisa_funds_prev_day_jpy", "nisa_funds_prev_day_pct", "nisa_funds_prev_month_jpy", "nisa_funds_prev_month_pct", "old_nisa_total_jpy", "old_nisa_prev_day_jpy", "old_nisa_prev_day_pct", "old_nisa_pnl_jpy", "old_nisa_pnl_pct", "cash_jpy_amount", "cash_jpy_value_jpy", "cash_usd_amount", "cash_usd_value_jpy", "other_funds_amount", "other_funds_value_jpy"}
	args := make([]any, len(columns))
	args[0] = "2026-08-16 12:00:00"
	args[1] = "OK"
	args[2] = 1
	for index := 3; index < len(args); index++ {
		args[index] = 0
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(columns)), ",")
	result, err := db.Exec("INSERT INTO sbi_snapshot ("+strings.Join(columns, ",")+") VALUES ("+placeholders+")", args...)
	if err != nil {
		t.Fatal("insert legacy snapshot")
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal("read legacy snapshot id")
	}
	return id
}

func insertLegacySbiHolding(t *testing.T, db *sql.DB, snapshotID int64) {
	t.Helper()
	_, err := db.Exec("INSERT INTO sbi_holding (snapshot_id, section, name, quantity, unit_cost, unit_price, prev_day_jpy, prev_day_pct, pnl_jpy, pnl_pct, value_jpy) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", snapshotID, "nisa_domestic", "ダミー銘柄A", 1, 1, 1, 0, 0, 0, 0, 1)
	if err != nil {
		t.Fatal("insert legacy holding")
	}
}

func insertLegacySbiHoldingWithFIGI(t *testing.T, db *sql.DB, snapshotID int64, figi string) {
	t.Helper()
	if _, err := db.Exec("INSERT INTO sbi_holding (snapshot_id, section, composite_figi, name, quantity, unit_cost, unit_price, prev_day_jpy, prev_day_pct, pnl_jpy, pnl_pct, value_jpy) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", snapshotID, "nisa_domestic", figi, "ダミー銘柄A", 1, 1, 1, 0, 0, 0, 0, 1); err != nil {
		t.Fatal("insert FIGI holding")
	}
}
