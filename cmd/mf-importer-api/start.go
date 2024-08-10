package main

import (
	"context"
	"fmt"
	"mf-importer/internal/logger"
	"mf-importer/internal/server"

	"github.com/spf13/cobra"
)

var dryRun bool
var inputDir string

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
	startCmd.Flags().StringVarP(&inputDir, "input-dir", "d", "/data/", "input directory")
}

func startMain() error {
	l := logger.NewLogger()
	ctx := context.Background()
	// host := os.Getenv("DB_HOST")
	// port := os.Getenv("DB_PORT")
	// user := os.Getenv("DB_USER")
	// pass := os.Getenv("DB_PASS")
	// name := os.Getenv("DB_NAME")

	// if host == "" {
	// 	host = "127.0.0.1"
	// }
	// if port == "" {
	// 	port = "3306"
	// }
	// if user == "" {
	// 	user = "root"
	// }
	// if pass == "" {
	// 	pass = "password"
	// }
	// if name == "" {
	// 	name = "mfimporter"
	// }

	// l.Info("using DB info",
	// 	zap.String("db_host", host),
	// 	zap.String("db_port", port),
	// 	zap.String("db_user", user),
	// 	zap.String("db_name", name),
	// )
	// db, err := repository.NewDBRepository(
	// 	host,
	// 	port,
	// 	user,
	// 	pass,
	// 	name,
	// )
	// if err != nil {
	// 	l.Error("failed to connect DB", zap.Error(err))
	// 	return err
	// }
	// defer db.CloseDB()

	server := server.Server{Logger: l}
	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}
