package repository

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	driverMySQL "github.com/go-sql-driver/mysql"
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

func expectSbiSnapshotLookup(mock sqlmock.Sqlmock, exists bool) {
	rows := sqlmock.NewRows([]string{"id"})
	if exists {
		rows.AddRow(41)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `sbi_snapshot` WHERE fetched_at = ? LIMIT ?")).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(rows)
}

func TestDBClient_ImportSbiSnapshot_New(t *testing.T) {
	db, mock := newSbiSQLMock(t)
	snapshot := &model.SbiSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.SbiStatusOK,
	}
	holdings := []model.SbiHolding{{Name: "ダミー銘柄A"}, {Name: "ダミー銘柄B"}}

	expectSbiSnapshotLookup(mock, false)
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

	expectSbiSnapshotLookup(mock, true)

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

func TestDBClient_ImportSbiSnapshot_DuplicateRaceIsSkipped(t *testing.T) {
	db, mock := newSbiSQLMock(t)
	snapshot := &model.SbiSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.SbiStatusOK,
	}
	holdings := []model.SbiHolding{{Name: "ダミー銘柄A"}}
	duplicateErr := &driverMySQL.MySQLError{Number: 1062, Message: "Duplicate entry for key 'uq_fetched_at'"}

	expectSbiSnapshotLookup(mock, false)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_snapshot`")).WillReturnError(duplicateErr)
	mock.ExpectRollback()
	expectSbiSnapshotLookup(mock, true)

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

func TestDBClient_ImportSbiSnapshot_DuplicateRaceWithoutSnapshotPropagates(t *testing.T) {
	db, mock := newSbiSQLMock(t)
	snapshot := &model.SbiSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.SbiStatusOK,
	}
	duplicateErr := &driverMySQL.MySQLError{Number: 1062, Message: "Duplicate entry for key 'uq_fetched_at'"}

	expectSbiSnapshotLookup(mock, false)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_snapshot`")).WillReturnError(duplicateErr)
	mock.ExpectRollback()
	expectSbiSnapshotLookup(mock, false)

	inserted, err := db.ImportSbiSnapshot(t.Context(), snapshot, nil)
	if inserted || err == nil {
		t.Fatalf("result = inserted %v, error %v; want propagated failure", inserted, err)
	}
	if !errors.Is(err, duplicateErr) {
		t.Fatal("database cause was not preserved")
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

	expectSbiSnapshotLookup(mock, false)
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

func TestDBClient_ImportSbiSnapshot_SanitizesDatabaseErrors(t *testing.T) {
	db, mock := newSbiSQLMock(t)
	snapshot := &model.SbiSnapshot{
		FetchedAt:     time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:        model.SbiStatusOK,
		SchemaVersion: model.CurrentSbiSchemaVersion,
	}
	figi := "DUMMY0000001"
	const holdingName = "ダミー銘柄A"
	holdings := []model.SbiHolding{{Name: holdingName, CompositeFIGI: &figi}}
	driverErr := &driverMySQL.MySQLError{
		Number:  1062,
		Message: "Duplicate entry '" + figi + "-" + holdingName + "' for key 'uq_snapshot_section_figi'",
	}

	expectSbiSnapshotLookup(mock, false)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_snapshot`")).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `sbi_holding`")).WillReturnError(driverErr)
	mock.ExpectRollback()

	inserted, err := db.ImportSbiSnapshot(t.Context(), snapshot, holdings)
	if inserted || err == nil {
		t.Fatalf("result = inserted %v, error %v; want sanitized failure", inserted, err)
	}
	if errors.Is(err, driverErr) == false {
		t.Fatal("database cause was not preserved")
	}
	var gotDriverErr *driverMySQL.MySQLError
	if !errors.As(err, &gotDriverErr) || gotDriverErr != driverErr {
		t.Fatal("database error was not available through errors.As")
	}
	if strings.Contains(err.Error(), figi) || strings.Contains(err.Error(), holdingName) {
		t.Fatalf("error echoed holding data: %q", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}
