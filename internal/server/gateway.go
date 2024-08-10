package server

import (
	"context"
	"encoding/json"
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
	var defaultLimit = 50
	if params.Limit == nil {
		params.Limit = &defaultLimit
	}

	dets, err := a.APIService.GetDetails(r.Context(), *params.Limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	outputJson, err := json.Marshal(&dets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(outputJson))
}
