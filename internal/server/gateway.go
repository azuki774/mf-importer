package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
	"net/http"

	"go.uber.org/zap"
)

const patchOpeReset = "reset"

type APIService interface {
	GetDetails(ctx context.Context, limit int, offset int, sort string, order string) (dets []openapi.Detail, err error)
	GetDetailsCount(ctx context.Context) (openapi.DetailsCount, error)
	GetRules(ctx context.Context, sort string, order string) ([]openapi.Rule, error)
	GetRule(ctx context.Context, id int) (openapi.Rule, error)
	ResetImportDetails(ctx context.Context, id int) (err error)
	AddRule(ctx context.Context, req openapi.RuleRequest) (openapi.Rule, error)
	DeleteRule(ctx context.Context, id int) error
	ListFinancialAssetSnapshots(context.Context, openapi.ListFinancialAssetSnapshotsParams) (openapi.SnapshotPage, error)
	GetFinancialAssetSnapshot(context.Context, string) (openapi.SnapshotDetail, error)
	GetFinancialAssetBalances(context.Context, openapi.GetFinancialAssetBalancesParams) (openapi.BalancePage, error)
}

func (a *apigateway) ListFinancialAssetSnapshots(w http.ResponseWriter, r *http.Request, p openapi.ListFinancialAssetSnapshotsParams) {
	v, err := a.APIService.ListFinancialAssetSnapshots(r.Context(), p)
	a.writeFinancialAsset(w, v, err)
}
func (a *apigateway) GetFinancialAssetSnapshot(w http.ResponseWriter, r *http.Request, id string) {
	v, err := a.APIService.GetFinancialAssetSnapshot(r.Context(), id)
	a.writeFinancialAsset(w, v, err)
}
func (a *apigateway) GetFinancialAssetBalances(w http.ResponseWriter, r *http.Request, p openapi.GetFinancialAssetBalancesParams) {
	v, err := a.APIService.GetFinancialAssetBalances(r.Context(), p)
	a.writeFinancialAsset(w, v, err)
}
func (a *apigateway) writeFinancialAsset(w http.ResponseWriter, v any, err error) {
	if err != nil {
		code, msg := http.StatusInternalServerError, "internal server error"
		if errors.Is(err, model.ErrInvalidFinancialAssetRequest) {
			code, msg = http.StatusBadRequest, "invalid financial asset request"
		} else if errors.Is(err, model.ErrRecordNotFound) {
			code, msg = http.StatusNotFound, "financial asset snapshot not found"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(openapi.ApiError{Error: msg})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if e := json.NewEncoder(w).Encode(v); e != nil {
		return
	}
}

type apigateway struct {
	openapi.Unimplemented
	Logger     *zap.Logger
	APIService APIService
}

// (GET /details)
func (a *apigateway) GetHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK\n")
}

// (GET /details)
func (a *apigateway) GetDetails(w http.ResponseWriter, r *http.Request, params openapi.GetDetailsParams) {
	var defaultLimit = 20
	var defaultOffset = 0
	var defaultSort = "useDate"
	var defaultOrder = "desc"
	if params.Limit == nil {
		params.Limit = &defaultLimit
	}
	if params.Offset == nil {
		params.Offset = &defaultOffset
	}
	if params.Sort == nil || *params.Sort == "" {
		params.Sort = &defaultSort
	}
	if params.Order == nil || *params.Order == "" {
		params.Order = &defaultOrder
	}

	dets, err := a.APIService.GetDetails(r.Context(), *params.Limit, *params.Offset, *params.Sort, *params.Order)
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

// (GET /details/count)
func (a *apigateway) GetDetailsCount(w http.ResponseWriter, r *http.Request) {
	res, err := a.APIService.GetDetailsCount(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	outputJson, err := json.Marshal(&res)
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
	if params.Ope == patchOpeReset {
		if err := a.APIService.ResetImportDetails(r.Context(), id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("not defined operation"))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// (GET /histories)
func (a *apigateway) GetHistories(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// (GET /rules)
func (a *apigateway) GetRules(w http.ResponseWriter, r *http.Request, params openapi.GetRulesParams) {
	var defaultSort = "id"
	var defaultOrder = "asc"
	if params.Sort == nil || *params.Sort == "" {
		params.Sort = &defaultSort
	}
	if params.Order == nil || *params.Order == "" {
		params.Order = &defaultOrder
	}
	rules, err := a.APIService.GetRules(r.Context(), *params.Sort, *params.Order)
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

// (GET /rules/{id})
func (a *apigateway) DeleteRulesId(w http.ResponseWriter, r *http.Request, id int) {
	err := a.APIService.DeleteRule(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// (GET /rules/{id})
func (a *apigateway) GetRulesId(w http.ResponseWriter, r *http.Request, id int) {
	rule, err := a.APIService.GetRule(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
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
