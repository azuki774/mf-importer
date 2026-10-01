package server

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"mf-importer/internal/openapi"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	Logger     *zap.Logger
	APIService APIService
	StaticDir  string // 指定時はビルド済みフロントエンドを配信する
	Addr       string // 未指定なら ":8080"
}

func (s *Server) Start(ctx context.Context) error {
	swagger, err := openapi.GetSwagger()
	if err != nil {
		s.Logger.Error("failed to get swagger spec", zap.Error(err))
		return err
	}
	swagger.Servers = nil
	r := chi.NewRouter()

	gw := &apigateway{Logger: s.Logger, APIService: s.APIService}

	if s.StaticDir != "" {
		// 静的配信時は API を /api 配下に限定する (フロントの相対パス呼び出しに合わせる)
		r.Route("/api", func(sub chi.Router) {
			registerAPI(gw, sub)
		})
		r.Handle("/*", newSPAHandler(s.StaticDir))
		s.Logger.Info("serve static frontend", zap.String("dir", s.StaticDir))
	} else {
		registerAPI(gw, r)
	}

	addr := s.Addr
	if addr == "" {
		addr = ":8080"
	}
	if err := http.ListenAndServe(addr, r); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			s.Logger.Error("failed to listen and serve", zap.Error(err))
			return err
		}
		// ErrServerClosed
	}

	return nil
}

func registerAPI(gw *apigateway, r chi.Router) {
	validateRaw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			p := strings.TrimPrefix(req.URL.Path, "/api")
			if strings.HasPrefix(p, "/v2/financial-assets/") {
				allowed := map[string]bool{}
				if p == "/v2/financial-assets/snapshots" || p == "/v2/financial-assets/balances" {
					allowed = map[string]bool{"source": true, "from": true, "to": true, "limit": true, "offset": true}
				}
				if p == "/v2/financial-assets/balances" {
					allowed["at"], allowed["interval"] = true, true
				}
				query, err := url.ParseQuery(req.URL.RawQuery)
				if err != nil {
					writeBadRequest(w)
					return
				}
				for k, vals := range query {
					if !allowed[k] || (k != "source" && len(vals) != 1) || len(vals) == 0 {
						writeBadRequest(w)
						return
					}
					for _, val := range vals {
						if val == "" {
							writeBadRequest(w)
							return
						}
					}
					if k == "source" {
						src := map[string]bool{}
						for _, v := range vals {
							if (v != "sbi" && v != "nrkn") || src[v] {
								writeBadRequest(w)
								return
							}
							src[v] = true
						}
					}
				}
			}
			next.ServeHTTP(w, req)
		})
	}
	openapi.HandlerWithOptions(gw, openapi.ChiServerOptions{BaseRouter: r, Middlewares: []openapi.MiddlewareFunc{validateRaw}, ErrorHandlerFunc: func(w http.ResponseWriter, req *http.Request, err error) {
		if strings.HasPrefix(req.URL.Path, "/v2/financial-assets/") || strings.HasPrefix(req.URL.Path, "/api/v2/financial-assets/") {
			writeBadRequest(w)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
	}})
}

func writeBadRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(openapi.ApiError{Error: "invalid request parameters"})
}

// newSPAHandler は静的ファイルを配信し、存在しないパスは index.html へフォールバックする
func newSPAHandler(dir string) http.HandlerFunc {
	fsys := os.DirFS(dir)
	return func(w http.ResponseWriter, r *http.Request) {
		upath := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if upath == "" || upath == "." {
			upath = "index.html"
		}
		info, err := fs.Stat(fsys, upath)
		if err != nil || info.IsDir() {
			upath = "index.html"
		}
		http.ServeFileFS(w, r, fsys, upath)
	}
}
