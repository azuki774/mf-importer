package main

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/metrics"
	"mf-importer/internal/repository"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

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
		l := logger.NewLogger()
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASS")
		name := os.Getenv("DB_NAME")

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
			zap.String("DB_USER", user),
			zap.String("DB_PASS", name),
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

		srv := metrics.MetricsServer{
			Logger:   l,
			Port:     "9000",
			JobLabel: os.Getenv("job_label"),
			DBClient: db,
		}
		if err := srv.Start(ctx); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
