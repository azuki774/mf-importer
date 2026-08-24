package server

import (
	"context"
	"errors"
	"io/fs"
	"mf-importer/internal/openapi"
	"net/http"
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
			openapi.HandlerFromMux(gw, sub)
		})
		r.Handle("/*", newSPAHandler(s.StaticDir))
		s.Logger.Info("serve static frontend", zap.String("dir", s.StaticDir))
	} else {
		openapi.HandlerFromMux(gw, r)
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
