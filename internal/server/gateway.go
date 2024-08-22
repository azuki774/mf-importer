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
	GetRules(ctx context.Context) ([]openapi.Rule, error)
	AddRule(ctx context.Context, req openapi.RuleRequest) (openapi.Rule, error)
}

type apigateway struct {
	Logger     *zap.Logger
	APIService APIService
}

// (GET /details)
func (a *apigateway) GetHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

// (GET /details/{id})
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

// (DELETE /details/{id})
func (a *apigateway) DeleteDetailsId(w http.ResponseWriter, r *http.Request, id int) {
	// TODO
}

// (GET /details/{id})
func (a *apigateway) GetDetailsId(w http.ResponseWriter, r *http.Request, id int) {
	// TODO
}

// (PATCH /details/{id})
func (a *apigateway) PatchDetailsId(w http.ResponseWriter, r *http.Request, id int, params openapi.PatchDetailsIdParams) {
	// TODO
}

// (GET /histories)
func (a *apigateway) GetHistories(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// (GET /rules)
func (a *apigateway) GetRules(w http.ResponseWriter, r *http.Request) {
	rules, err := a.APIService.GetRules(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	outputJson, err := json.Marshal(&rules)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(outputJson))
}

// (POST /rules)
func (a *apigateway) PostRules(w http.ResponseWriter, r *http.Request) {
	var req openapi.RuleRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err.Error())
		return
	}
	rule, err := a.APIService.AddRule(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	outputJson, err := json.Marshal(&rule)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(outputJson))
}
