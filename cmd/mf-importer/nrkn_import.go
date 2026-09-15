package main

import (
	"context"
	"fmt"
	"mf-importer/internal/logger"
	"mf-importer/internal/repository"
	"mf-importer/internal/service"
	"os"
	"path/filepath"
	"time"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type nrknMode string

const (
	nrknLocalMode nrknMode = "local"
	nrknS3Mode    nrknMode = "s3"
)

var (
	nrknInputDir string
	nrknMonth    string
)

var nrknImportCmd = &cobra.Command{
	Use:   "nrkn-import",
	Short: "import NRKN snapshots into the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNrknImport()
	},
}

func init() {
	rootCmd.AddCommand(nrknImportCmd)
	nrknImportCmd.Flags().StringVar(&nrknInputDir, "input-dir", "", "read NRKN JSON files from a local directory instead of S3")
	nrknImportCmd.Flags().StringVar(&nrknMonth, "month", "", "S3 target month in YYYYMM form (defaults to the current Tokyo month)")
}

var nrknTokyo = time.FixedZone("Asia/Tokyo", 9*60*60)

func nrknImportMode(inputDir, month string) (nrknMode, error) {
	if inputDir != "" {
		if month != "" {
			return "", fmt.Errorf("--input-dir and --month cannot be used together")
		}
		return nrknLocalMode, nil
	}
	return nrknS3Mode, nil
}

func resolveNrknMonth(raw string, now time.Time) (string, error) {
	if raw == "" {
		return now.In(nrknTokyo).Format("200601"), nil
	}
	if len(raw) != 6 {
		return "", fmt.Errorf("invalid NRKN month %q: want YYYYMM", raw)
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("invalid NRKN month %q: want YYYYMM", raw)
		}
	}
	if _, err := time.ParseInLocation("200601", raw, nrknTokyo); err != nil {
		return "", fmt.Errorf("invalid NRKN month %q: want YYYYMM: %w", raw, err)
	}
	return raw, nil
}

func runNrknImport() error {
	l := logger.NewLogger()
	ctx := context.Background()
	mode, err := nrknImportMode(nrknInputDir, nrknMonth)
	if err != nil {
		return err
	}

	targetDir := nrknInputDir
	var month string
	if mode == nrknS3Mode {
		month, err = resolveNrknMonth(nrknMonth, time.Now())
		if err != nil {
			return err
		}
	} else {
		targetDir, err = filepath.Abs(targetDir)
		if err != nil {
			return fmt.Errorf("resolve NRKN input directory: %w", err)
		}
	}
	db, err := openImporterDatabase()
	if err != nil {
		return err
	}
	defer db.CloseDB()
	if err := applyMigrations(ctx, l, db, migrate.Up, 0); err != nil {
		return err
	}
	if mode == nrknS3Mode {
		targetDir, err = os.MkdirTemp("", "mf-importer-nrkn-")
		if err != nil {
			return fmt.Errorf("create NRKN staging directory: %w", err)
		}
		defer os.RemoveAll(targetDir)
		l.Info("start NRKN S3 download", zap.String("month", month))
		if err := repository.NewNrknDownloader(targetDir).StartMonth(ctx, month); err != nil {
			return err
		}
	}

	operator := &repository.NrknJSONOperator{Logger: l}
	importer := service.NewNrknImporter(l, db, operator)
	result, err := importer.Start(ctx, targetDir)
	if err != nil {
		return err
	}
	l.Info("NRKN import complete",
		zap.Int("files", result.Files),
		zap.Int("inserted", result.Inserted),
		zap.Int("skipped", result.Skipped),
		zap.Int("unsupported", result.Unsupported),
	)
	return nil
}
