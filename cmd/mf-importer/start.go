package main

import (
	"context"
	"fmt"
	"mf-importer/internal/logger"
	"mf-importer/internal/repository"
	"mf-importer/internal/service"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var dryRun bool
var withDownload bool // true ならば s3からのCSVダウンロードも
var inputDir string   // CSVダウンロード時に使うディレクトリを指定

// startCmd represents the regist command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("start called")
		return startMain()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// registCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().BoolVar(&dryRun, "dry-run", false, "dry run")
	startCmd.Flags().BoolVar(&withDownload, "with-download", false, "download and import")
	startCmd.Flags().StringVarP(&inputDir, "input-dir", "d", "/data/", "input directory")
}

func startMain() error {
	l := logger.NewLogger()
	ctx := context.Background()
	host := os.Getenv("db_host")
	port := os.Getenv("db_port")
	user := os.Getenv("db_user")
	pass := os.Getenv("db_pass")
	name := os.Getenv("db_name")

	if withDownload {
		l.Info("start download CSV from s3")
		downloader := repository.NewDownloader(inputDir)
		if err := downloader.Start(ctx); err != nil {
			l.Error("failed to download", zap.Error(err))
			return err
		}
		l.Info("complete download CSV from s3")
	}

	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "3306"
	}
	if user == "" {
		user = "root"
	}
	if pass == "" {
		pass = "password"
	}
	if name == "" {
		name = "mfimporter"
	}

	l.Info("using DB info",
		zap.String("DB_HOST", host),
		zap.String("DB_PORT", port),
		zap.String("DB_NAME", name),
		zap.String("DB_USER", user),
		zap.String("DB_PASS", pass),
	)
	db, err := repository.NewDBRepository(
		host,
		port,
		user,
		pass,
		name,
	)
	if err != nil {
		l.Error("failed to connect DB", zap.Error(err))
		return err
	}
	defer db.CloseDB()

	importer := service.NewImporter(l, db, inputDir, dryRun)
	if err := importer.Start(ctx); err != nil {
		return err
	}

	return nil
}
