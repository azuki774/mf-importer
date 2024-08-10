package server

import (
	"context"
	"errors"
	"mf-importer/internal/openapi"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	Logger     *zap.Logger
	APIService APIService
}

func (s *Server) Start(ctx context.Context) error {
	swagger, err := openapi.GetSwagger()
	if err != nil {
		s.Logger.Error("failed to get swagger spec", zap.Error(err))
		return err
	}
	swagger.Servers = nil
	r := chi.NewRouter()

	openapi.HandlerFromMux(&apigateway{Logger: s.Logger}, r)
	addr := ":8080"
	if err := http.ListenAndServe(addr, r); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			s.Logger.Error("failed to listen and serve", zap.Error(err))
			return err
		}
		// ErrServerClosed
	}

	return nil
}
