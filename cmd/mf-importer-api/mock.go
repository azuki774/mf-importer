package main

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/mfapi"
	"mf-importer/internal/server"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var mockInputDir string
var mockStaticDir string
var mockAddr string

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "start mock API server with fixture data (no DB required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mockMain()
	},
}

func init() {
	rootCmd.AddCommand(mockCmd)
	mockCmd.Flags().StringVarP(&mockInputDir, "input-dir", "d", "./test/", "directory of CSV fixtures")
	mockCmd.Flags().StringVar(&mockStaticDir, "static-dir", "", "serve built frontend from this directory")
	mockCmd.Flags().StringVar(&mockAddr, "addr", ":8080", "listen address")
}

func mockMain() error {
	l := logger.NewLogger()
	ctx := context.Background()

	svc, err := mfapi.NewMockAPIService(l, mockInputDir)
	if err != nil {
		l.Error("failed to build mock fixtures", zap.Error(err))
		return err
	}
	l.Info("mock fixtures loaded",
		zap.String("input_dir", mockInputDir),
		zap.String("static_dir", mockStaticDir),
	)

	srv := server.Server{Logger: l, APIService: svc, StaticDir: mockStaticDir, Addr: mockAddr}
	return srv.Start(ctx)
}
