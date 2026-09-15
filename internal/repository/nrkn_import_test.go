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
	"gorm.io/gorm/logger"
	"mf-importer/internal/model"
)

func newNrknSQLMock(t *testing.T) (*DBClient, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return &DBClient{Conn: gormDB}, mock
}

func expectNrknSnapshotLookup(mock sqlmock.Sqlmock, exists bool) {
	rows := sqlmock.NewRows([]string{"id"})
	if exists {
		rows.AddRow(41)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT `id` FROM `nrkn_snapshot` WHERE fetched_at = ? LIMIT ?")).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(rows)
}

func TestDBClient_ImportNrknSnapshot_New(t *testing.T) {
	db, mock := newNrknSQLMock(t)
	snapshot := &model.NrknSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.NrknStatusOK,
	}
	holdings := []model.NrknHolding{{Name: "ダミー銘柄A"}, {Name: "ダミー銘柄B"}}

	expectNrknSnapshotLookup(mock, false)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `nrkn_snapshot`")).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `nrkn_holding`")).WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	inserted, err := db.ImportNrknSnapshot(t.Context(), snapshot, holdings)
	if err != nil {
		t.Fatalf("ImportNrknSnapshot: %v", err)
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

func TestDBClient_ImportNrknSnapshot_DuplicateIsSkipped(t *testing.T) {
	db, mock := newNrknSQLMock(t)
	snapshot := &model.NrknSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.NrknStatusOK,
	}
	holdings := []model.NrknHolding{{Name: "ダミー銘柄A"}}

	expectNrknSnapshotLookup(mock, true)

	inserted, err := db.ImportNrknSnapshot(t.Context(), snapshot, holdings)
	if err != nil {
		t.Fatalf("ImportNrknSnapshot: %v", err)
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

func TestDBClient_ImportNrknSnapshot_DuplicateRaceIsSkipped(t *testing.T) {
	db, mock := newNrknSQLMock(t)
	snapshot := &model.NrknSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.NrknStatusOK,
	}
	holdings := []model.NrknHolding{{Name: "ダミー銘柄A"}}
	duplicateErr := &driverMySQL.MySQLError{Number: 1062, Message: "Duplicate entry for key 'uq_fetched_at'"}

	expectNrknSnapshotLookup(mock, false)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `nrkn_snapshot`")).WillReturnError(duplicateErr)
	mock.ExpectRollback()
	expectNrknSnapshotLookup(mock, true)

	inserted, err := db.ImportNrknSnapshot(t.Context(), snapshot, holdings)
	if err != nil {
		t.Fatalf("ImportNrknSnapshot: %v", err)
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

func TestDBClient_ImportNrknSnapshot_DuplicateRaceWithoutSnapshotPropagates(t *testing.T) {
	db, mock := newNrknSQLMock(t)
	snapshot := &model.NrknSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.NrknStatusOK,
	}
	duplicateErr := &driverMySQL.MySQLError{Number: 1062, Message: "Duplicate entry for key 'uq_fetched_at'"}

	expectNrknSnapshotLookup(mock, false)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `nrkn_snapshot`")).WillReturnError(duplicateErr)
	mock.ExpectRollback()
	expectNrknSnapshotLookup(mock, false)

	inserted, err := db.ImportNrknSnapshot(t.Context(), snapshot, nil)
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

func TestDBClient_ImportNrknSnapshot_RollsBackHoldingFailure(t *testing.T) {
	db, mock := newNrknSQLMock(t)
	snapshot := &model.NrknSnapshot{
		FetchedAt: time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:    model.NrknStatusOK,
	}
	holdings := []model.NrknHolding{{Name: "ダミー銘柄A"}}
	wantErr := errors.New("synthetic holding insert failure")

	expectNrknSnapshotLookup(mock, false)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `nrkn_snapshot`")).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `nrkn_holding`")).WillReturnError(wantErr)
	mock.ExpectRollback()

	inserted, err := db.ImportNrknSnapshot(t.Context(), snapshot, holdings)
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

func TestDBClient_ImportNrknSnapshot_SanitizesDatabaseErrors(t *testing.T) {
	db, mock := newNrknSQLMock(t)
	snapshot := &model.NrknSnapshot{
		FetchedAt:     time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC),
		Status:        model.NrknStatusOK,
		SchemaVersion: model.CurrentNrknSchemaVersion,
	}
	figi := "DUMMY0000001"
	const holdingName = "ダミー銘柄A"
	holdings := []model.NrknHolding{{Name: holdingName, CompositeFIGI: figi}}
	driverErr := &driverMySQL.MySQLError{
		Number:  1062,
		Message: "Duplicate entry '" + figi + "-" + holdingName + "' for key 'uq_snapshot_section_figi'",
	}

	expectNrknSnapshotLookup(mock, false)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `nrkn_snapshot`")).WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `nrkn_holding`")).WillReturnError(driverErr)
	mock.ExpectRollback()

	inserted, err := db.ImportNrknSnapshot(t.Context(), snapshot, holdings)
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
