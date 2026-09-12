package migration

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysql "github.com/go-sql-driver/mysql"
	migrate "github.com/rubenv/sql-migrate"
	assets "mf-importer/migration"
)

func TestEmbeddedMigrations(t *testing.T) {
	source := migrate.EmbedFileSystemMigrationSource{FileSystem: assets.Files, Root: "db"}
	migrations, err := source.FindMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) != 7 {
		t.Fatalf("got %d migrations", len(migrations))
	}
	for _, m := range migrations {
		if len(m.Up) == 0 || len(m.Down) == 0 {
			t.Errorf("missing direction in %s", m.Id)
		}
	}
}

func TestErrorDoesNotExposeDatabaseMessage(t *testing.T) {
	cause := &mysql.MySQLError{Number: 1064, Message: "synthetic-sensitive-value"}
	original := &migrate.TxError{Migration: &migrate.Migration{Id: "001_dummy.sql"}, Err: cause}
	err := NewError("application", original)
	if strings.Contains(err.Error(), cause.Message) {
		t.Fatal("driver message exposed")
	}
	if !strings.Contains(err.Error(), "001_dummy.sql") || !strings.Contains(err.Error(), "1064") {
		t.Fatal("missing safe diagnostics")
	}
	var dbErr *mysql.MySQLError
	if !errors.As(err, &dbErr) || dbErr != cause {
		t.Fatal("lost database cause")
	}
}

func TestLockFailureDoesNotApplySQL(t *testing.T) {
	for _, value := range []any{nil, int64(0)} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectQuery("SELECT DATABASE").WillReturnRows(sqlmock.NewRows([]string{"db"}).AddRow("dummy"))
		mock.ExpectQuery("SELECT GET_LOCK").WillReturnRows(sqlmock.NewRows([]string{"lock"}).AddRow(value))
		n, err := runLocked(context.Background(), db, &migrate.MemoryMigrationSource{}, migrate.Up, 0, 0)
		if err == nil || n != 0 {
			t.Fatal("expected lock failure")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
}

func TestSessionCannotReconnect(t *testing.T) {
	connector := &sessionConnector{}
	if _, err := connector.Connect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := connector.Connect(context.Background()); !errors.Is(err, sql.ErrConnDone) {
		t.Fatal("session could reconnect without its lock")
	}
}
