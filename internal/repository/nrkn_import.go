package repository

import (
	"context"
	"errors"
	"mf-importer/internal/model"
	"time"

	"gorm.io/gorm"
)

const tableNameNrknSnapshot = "nrkn_snapshot"
const tableNameNrknHolding = "nrkn_holding"

type nrknDatabaseError struct {
	operation string
	cause     error
}

func (e *nrknDatabaseError) Error() string {
	return "NRKN database operation failed: " + e.operation
}

func (e *nrknDatabaseError) Unwrap() error { return e.cause }

func newNrknDatabaseError(operation string, err error) error {
	return &nrknDatabaseError{operation: operation, cause: err}
}

// ImportNrknSnapshot atomically creates a snapshot and its holdings. A snapshot
// with an existing fetched_at is treated as an idempotent no-op.
func (d *DBClient) ImportNrknSnapshot(ctx context.Context, snapshot *model.NrknSnapshot, holdings []model.NrknHolding) (inserted bool, err error) {
	snapshot.FetchedAt = model.NormalizeNrknFetchedAt(snapshot.FetchedAt)
	exists, err := d.nrknSnapshotExists(ctx, snapshot.FetchedAt)
	if err != nil {
		return false, newNrknDatabaseError("check existing snapshot", err)
	}
	if exists {
		return false, nil
	}

	var duplicateErr error
	err = d.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Table(tableNameNrknSnapshot).Create(snapshot).Error; err != nil {
			if isMySQLDuplicateError(err) {
				duplicateErr = err
			}
			return newNrknDatabaseError("insert snapshot", err)
		}
		inserted = true

		for index := range holdings {
			holdings[index].SnapshotID = snapshot.ID
		}
		if len(holdings) == 0 {
			return nil
		}
		if err := tx.WithContext(ctx).Table(tableNameNrknHolding).Create(&holdings).Error; err != nil {
			return newNrknDatabaseError("insert holdings", err)
		}
		return nil
	})
	if err == nil {
		return inserted, nil
	}
	if duplicateErr == nil {
		return false, newNrknDatabaseError("transaction", err)
	}

	exists, checkErr := d.nrknSnapshotExists(ctx, snapshot.FetchedAt)
	if checkErr != nil {
		return false, newNrknDatabaseError("check existing snapshot after conflict", checkErr)
	}
	if exists {
		return false, nil
	}
	return false, err
}

func (d *DBClient) nrknSnapshotExists(ctx context.Context, fetchedAt time.Time) (bool, error) {
	var existing model.NrknSnapshot
	err := d.Conn.WithContext(ctx).
		Table(tableNameNrknSnapshot).
		Select("id").
		Where("fetched_at = ?", fetchedAt).
		Take(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
