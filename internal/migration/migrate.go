// Package migration applies the embedded sql-migrate migrations.
package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	assets "mf-importer/migration"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	migrate "github.com/rubenv/sql-migrate"
)

const lockWaitSeconds = 60

// Error keeps the database cause available to errors.As without printing SQL,
// credentials, or data that may appear in a driver error message.
type Error struct {
	Stage       string
	MigrationID string
	cause       error
}

func (e *Error) Error() string {
	message := "migration " + e.Stage + " failed"
	if e.MigrationID != "" {
		message += " (" + e.MigrationID + ")"
	}
	var dbErr *mysql.MySQLError
	if errors.As(e.cause, &dbErr) {
		message += fmt.Sprintf(" [mysql error %d]", dbErr.Number)
	}
	return message
}
func (e *Error) Unwrap() error { return e.cause }

// NewError wraps a cause without exposing its potentially sensitive message.
func NewError(stage string, err error) error {
	e := &Error{Stage: stage, cause: err}
	var txErr *migrate.TxError
	if errors.As(err, &txErr) {
		e.MigrationID = txErr.Migration.Id
		e.cause = errors.Join(err, txErr.Err)
	}
	return e
}

// Run applies Up or Down on a dedicated session. A zero limit means all.
// The caller owns db; this function never changes its pool configuration.
func Run(ctx context.Context, db *sql.DB, direction migrate.MigrationDirection, limit int) (int, error) {
	source := &migrate.EmbedFileSystemMigrationSource{FileSystem: assets.Files, Root: "db"}
	return run(ctx, db, source, direction, limit, lockWaitSeconds)
}

func run(ctx context.Context, db *sql.DB, source migrate.MigrationSource, direction migrate.MigrationDirection, limit, wait int) (count int, err error) {
	if limit < 0 {
		return 0, errors.New("migration limit must be non-negative")
	}
	if direction != migrate.Up && direction != migrate.Down {
		return 0, errors.New("invalid migration direction")
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return 0, NewError("connect", err)
	}
	defer conn.Close()

	// sql-migrate accepts *sql.DB, while GET_LOCK belongs to a physical session.
	// Adapt the reserved connection so the lock, DDL and history writes use the
	// same session. Reconnection is forbidden: losing the session loses the lock.
	err = conn.Raw(func(raw any) error {
		physical, ok := raw.(migrationConn)
		if !ok {
			return NewError("connect", errors.New("unsupported database driver"))
		}
		session := sql.OpenDB(&sessionConnector{conn: physical})
		session.SetMaxOpenConns(1)
		session.SetMaxIdleConns(1)
		defer session.Close()
		count, err = runLocked(ctx, session, source, direction, limit, wait)
		return err
	})
	if err != nil {
		// Discard the reserved connection after any failure, including an uncertain
		// RELEASE_LOCK result. Conn.Close alone would return it to the caller's pool.
		_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		var safe *Error
		if !errors.As(err, &safe) {
			err = NewError("session", err)
		}
	}
	return count, err
}

func runLocked(ctx context.Context, db *sql.DB, source migrate.MigrationSource, direction migrate.MigrationDirection, limit, wait int) (count int, err error) {
	var name string
	if err := db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&name); err != nil {
		return 0, NewError("database selection", err)
	}
	digest := sha256.Sum256([]byte(name))
	lockName := fmt.Sprintf("mf-importer:migrate:%x", digest[:16])
	var acquired sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", lockName, wait).Scan(&acquired); err != nil {
		return 0, NewError("lock acquisition", err)
	}
	if !acquired.Valid || acquired.Int64 != 1 {
		return 0, NewError("lock acquisition", errors.New("lock unavailable or timed out"))
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var released sql.NullInt64
		releaseErr := db.QueryRowContext(releaseCtx, "SELECT RELEASE_LOCK(?)", lockName).Scan(&released)
		if releaseErr == nil && (!released.Valid || released.Int64 != 1) {
			releaseErr = errors.New("lock no longer owned")
		}
		if releaseErr != nil {
			err = errors.Join(err, NewError("lock release", releaseErr))
		}
	}()
	set := migrate.MigrationSet{TableName: "gorp_migrations"}
	count, err = set.ExecMaxContext(ctx, db, "mysql", source, direction, limit)
	if err != nil {
		return count, NewError("application", err)
	}
	return count, nil
}

// Preserve the MySQL driver's context-aware operations through the adapter.
type migrationConn interface {
	driver.Conn
	driver.ConnBeginTx
	driver.ExecerContext
	driver.QueryerContext
}

type sessionConn struct{ migrationConn }

// The outer *sql.Conn owns the physical connection.
func (*sessionConn) Close() error { return nil }

type sessionConnector struct {
	conn migrationConn
	used bool
}

func (c *sessionConnector) Connect(context.Context) (driver.Conn, error) {
	if c.used {
		return nil, sql.ErrConnDone
	}
	c.used = true
	return &sessionConn{c.conn}, nil
}
func (*sessionConnector) Driver() driver.Driver { return &mysql.MySQLDriver{} }
