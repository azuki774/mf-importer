// Package migration contains the SQL shipped with this build.
package migration

import "embed"

// Files preserves the existing sql-migrate filenames and migration IDs.
//
//go:embed db/*.sql
var Files embed.FS
