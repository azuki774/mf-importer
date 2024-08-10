package server

import (
	"context"
	"fmt"
	"mf-importer/internal/openapi"
	"net/http"

	"go.uber.org/zap"
)

type APIService interface {
	GetDetails(ctx context.Context, limit int) (dets []openapi.Detail, err error)
}

type apigateway struct {
	Logger     *zap.Logger
	APIService APIService
}

func (a *apigateway) GetHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

func (a *apigateway) GetDetails(w http.ResponseWriter, r *http.Request, params openapi.GetDetailsParams) {
	fmt.Fprintf(w, "Get GetDetails\n")
}
