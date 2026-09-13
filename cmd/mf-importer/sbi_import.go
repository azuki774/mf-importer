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

type sbiMode string

const (
	sbiLocalMode sbiMode = "local"
	sbiS3Mode    sbiMode = "s3"
)

var (
	sbiInputDir string
	sbiMonth    string
)

var sbiImportCmd = &cobra.Command{
	Use:   "sbi-import",
	Short: "import SBI snapshots into the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSbiImport()
	},
}

func init() {
	rootCmd.AddCommand(sbiImportCmd)
	sbiImportCmd.Flags().StringVar(&sbiInputDir, "input-dir", "", "read SBI JSON files from a local directory instead of S3")
	sbiImportCmd.Flags().StringVar(&sbiMonth, "month", "", "S3 target month in YYYYMM form (defaults to the current Tokyo month)")
}

var sbiTokyo = time.FixedZone("Asia/Tokyo", 9*60*60)

func sbiImportMode(inputDir, month string) (sbiMode, error) {
	if inputDir != "" {
		if month != "" {
			return "", fmt.Errorf("--input-dir and --month cannot be used together")
		}
		return sbiLocalMode, nil
	}
	return sbiS3Mode, nil
}

func resolveSbiMonth(raw string, now time.Time) (string, error) {
	if raw == "" {
		return now.In(sbiTokyo).Format("200601"), nil
	}
	if len(raw) != 6 {
		return "", fmt.Errorf("invalid SBI month %q: want YYYYMM", raw)
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("invalid SBI month %q: want YYYYMM", raw)
		}
	}
	if _, err := time.ParseInLocation("200601", raw, sbiTokyo); err != nil {
		return "", fmt.Errorf("invalid SBI month %q: want YYYYMM: %w", raw, err)
	}
	return raw, nil
}

func runSbiImport() error {
	l := logger.NewLogger()
	ctx := context.Background()
	mode, err := sbiImportMode(sbiInputDir, sbiMonth)
	if err != nil {
		return err
	}

	targetDir := sbiInputDir
	var month string
	if mode == sbiS3Mode {
		month, err = resolveSbiMonth(sbiMonth, time.Now())
		if err != nil {
			return err
		}
	} else {
		targetDir, err = filepath.Abs(targetDir)
		if err != nil {
			return fmt.Errorf("resolve SBI input directory: %w", err)
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
	if mode == sbiS3Mode {
		targetDir, err = os.MkdirTemp("", "mf-importer-sbi-")
		if err != nil {
			return fmt.Errorf("create SBI staging directory: %w", err)
		}
		defer os.RemoveAll(targetDir)
		l.Info("start SBI S3 download", zap.String("month", month))
		if err := repository.NewSbiDownloader(targetDir).StartMonth(ctx, month); err != nil {
			return err
		}
	}

	operator := &repository.SbiJSONOperator{Logger: l}
	importer := service.NewSbiImporter(l, db, operator)
	result, err := importer.Start(ctx, targetDir)
	if err != nil {
		return err
	}
	l.Info("SBI import complete",
		zap.Int("files", result.Files),
		zap.Int("inserted", result.Inserted),
		zap.Int("skipped", result.Skipped),
	)
	return nil
}
