package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

func commandTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MF_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set MF_MIGRATION_TEST_DSN to a disposable MariaDB server")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid test DSN")
	}
	host, port, err := net.SplitHostPort(cfg.Addr)
	if err != nil || cfg.Net != "tcp" {
		t.Fatal("test DSN must use TCP")
	}
	cfg.DBName = ""
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("open test server")
	}
	name := fmt.Sprintf("mf_command_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal("create test database")
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
	for key, value := range map[string]string{"DB_HOST": host, "DB_PORT": port, "DB_USER": cfg.User, "DB_PASS": cfg.Passwd, "DB_NAME": name} {
		t.Setenv(key, value)
	}
	return db
}

func TestIntegrationImporterStartup(t *testing.T) {
	db := commandTestDatabase(t)
	oldDry, oldDownload, oldDir := dryRun, withDownload, inputDir
	t.Cleanup(func() { dryRun, withDownload, inputDir = oldDry, oldDownload, oldDir })
	inputDir = t.TempDir()
	withDownload = false
	dryRun = true
	if err := startMain(); err != nil {
		t.Fatal(err)
	}
	var tables int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&tables); err != nil || tables != 0 {
		t.Fatal("dry run changed schema")
	}
	dryRun = false
	if err := startMain(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM gorp_migrations").Scan(&count); err != nil || count != 7 {
		t.Fatal("startup did not migrate")
	}
	// Migration must fail before accessing this missing input directory or S3.
	if _, err := db.Exec("INSERT INTO gorp_migrations (id, applied_at) VALUES ('999_dummy.sql', NOW())"); err != nil {
		t.Fatal("seed unknown history")
	}
	inputDir += "/missing"
	withDownload = true
	if err := startMain(); err == nil || !strings.Contains(err.Error(), "migration application") {
		t.Fatal("startup continued after migration failure", err)
	}
}

func TestIntegrationSbiStartup(t *testing.T) {
	db := commandTestDatabase(t)
	oldDir, oldMonth := sbiInputDir, sbiMonth
	t.Cleanup(func() { sbiInputDir, sbiMonth = oldDir, oldMonth })
	sbiInputDir = t.TempDir()
	sbiMonth = ""
	err := runSbiImport()
	if err == nil || !strings.Contains(err.Error(), "no SBI JSON files") {
		t.Fatal("expected empty input after migration", err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM gorp_migrations").Scan(&count); err != nil || count != 7 {
		t.Fatal("SBI startup did not migrate")
	}
	if _, err := db.Exec("INSERT INTO gorp_migrations (id, applied_at) VALUES ('999_dummy.sql', NOW())"); err != nil {
		t.Fatal("seed unknown history")
	}
	sbiInputDir = ""
	if err := runSbiImport(); err == nil || !strings.Contains(err.Error(), "migration application") {
		t.Fatal("SBI started downloading after migration failure", err)
	}
}

func TestMigrateRejectsNegativeLimit(t *testing.T) {
	command := newMigrateCommand()
	command.SetArgs([]string{"down", "--limit", "-1"})
	if err := command.ExecuteContext(context.Background()); err == nil {
		t.Fatal("accepted negative limit")
	}
}

func TestDatabaseConfigCompatibility(t *testing.T) {
	for _, key := range []string{"HOST", "PORT", "USER", "PASS", "NAME"} {
		t.Setenv("DB_"+key, "")
		t.Setenv("db_"+strings.ToLower(key), "legacy-dummy")
	}
	cfg := importerDatabaseConfig()
	if cfg.host != "legacy-dummy" || cfg.pass != "legacy-dummy" {
		t.Fatal("legacy environment ignored")
	}
	t.Setenv("DB_HOST", "primary-dummy")
	if importerDatabaseConfig().host != "primary-dummy" {
		t.Fatal("uppercase environment did not take precedence")
	}
}
