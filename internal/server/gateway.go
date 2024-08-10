package server

import (
	"fmt"
	"mf-importer/internal/openapi"
	"net/http"

	"go.uber.org/zap"
)

type apigateway struct {
	Logger *zap.Logger
}

func (a *apigateway) GetHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

func (a *apigateway) GetImports(w http.ResponseWriter, r *http.Request, params openapi.GetImportsParams) {
	fmt.Fprintf(w, "Get Imports\n")
}
