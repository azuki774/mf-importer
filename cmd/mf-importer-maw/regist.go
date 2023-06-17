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
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func registMain() error {
	l := logger.NewLogger()
	ctx := context.Background()
	l.Info("using DB uri", zap.String("db_uri", os.Getenv("db_uri")))
	db, err := repository.NewMongoDB(ctx, os.Getenv("db_uri"))
	if err != nil {
		l.Error("failed to connect DB", zap.Error(err))
		return err
	}
	defer db.Disconnect(ctx)

	mw := mawinter.NewMawinter(db)
	err = mw.Regist(ctx)
	if err != nil {
		return err
	}

	l.Info("registMain end")
	return nil
}
