package repository

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"mf-importer/internal/model"
)

func newSbiSQLMock(t *testing.T) (*DBClient, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return &DBClient{Conn: gormDB}, mock
}

func TestDBClient_ImportSbiSnapshot_New(t *testing.T) {
	db, mock := newSbiSQLMock(t)
	snapshot := &model.SbiSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.SbiStatusOK,
	}
	holdings := []model.SbiHolding{{Name: "ダミー銘柄A"}, {Name: "ダミー銘柄B"}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_snapshot`")).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_holding`")).WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	inserted, err := db.ImportSbiSnapshot(t.Context(), snapshot, holdings)
	if err != nil {
		t.Fatalf("ImportSbiSnapshot: %v", err)
	}
	if !inserted {
		t.Fatal("inserted = false, want true")
	}
	if snapshot.ID != 42 {
		t.Fatalf("snapshot ID = %d, want 42", snapshot.ID)
	}
	if holdings[0].SnapshotID != 42 || holdings[1].SnapshotID != 42 {
		t.Fatalf("holding snapshot IDs = %d, %d, want 42, 42", holdings[0].SnapshotID, holdings[1].SnapshotID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestDBClient_ImportSbiSnapshot_DuplicateIsSkipped(t *testing.T) {
	db, mock := newSbiSQLMock(t)
	snapshot := &model.SbiSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.SbiStatusOK,
	}
	holdings := []model.SbiHolding{{Name: "ダミー銘柄A"}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_snapshot`")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	inserted, err := db.ImportSbiSnapshot(t.Context(), snapshot, holdings)
	if err != nil {
		t.Fatalf("ImportSbiSnapshot: %v", err)
	}
	if inserted {
		t.Fatal("inserted = true, want false")
	}
	if holdings[0].SnapshotID != 0 {
		t.Fatalf("holding snapshot ID = %d, want unchanged", holdings[0].SnapshotID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestDBClient_ImportSbiSnapshot_RollsBackHoldingFailure(t *testing.T) {
	db, mock := newSbiSQLMock(t)
	snapshot := &model.SbiSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.SbiStatusOK,
	}
	holdings := []model.SbiHolding{{Name: "ダミー銘柄A"}}
	wantErr := errors.New("synthetic holding insert failure")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_snapshot`")).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_holding`")).WillReturnError(wantErr)
	mock.ExpectRollback()

	inserted, err := db.ImportSbiSnapshot(t.Context(), snapshot, holdings)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if inserted {
		t.Fatal("inserted = true, want false on rollback")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}
