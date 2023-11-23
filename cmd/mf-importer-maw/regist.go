package main

import (
	"context"
	"fmt"
	"mf-importer/internal/logger"
	"mf-importer/internal/mawinter"
	"mf-importer/internal/repository"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var dryRun bool

// registCmd represents the regist command
var registCmd = &cobra.Command{
	Use:   "regist",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("regist called")
		return registMain()
	},
}

func init() {
	rootCmd.AddCommand(registCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// registCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	registCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "dry run mode")
}

func registMain() error {
	l := logger.NewLogger()
	ctx := context.Background()
	host := os.Getenv("db_host")
	port := os.Getenv("db_port")
	user := os.Getenv("db_user")
	pass := os.Getenv("db_pass")
	name := os.Getenv("db_name")

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
		zap.String("db_host", host),
		zap.String("db_port", port),
		zap.String("db_user", user),
		zap.String("db_name", name),
	)
	l.Info("using mawinter API post endpoint", zap.String("api_uri", os.Getenv("api_uri")), zap.String("get_uri", os.Getenv("get_uri")))
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

	maw := repository.NewMawinterClient(os.Getenv("api_uri"), os.Getenv("get_uri"))
	mw := mawinter.NewMawinter(db, maw, dryRun)

	err = mw.Regist(ctx)
	if err != nil {
		return err
	}

	return nil
}
