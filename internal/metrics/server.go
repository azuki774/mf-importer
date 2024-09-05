package metrics

import (
	"context"
	"errors"
	"fmt"
	"mf-importer/internal/model"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type DBClient interface {
	GetLastDetailHistoryWhereJobLabel(ctx context.Context, jobLabel string) (model.ImportHistory, error)
}

type MetricsServer struct {
	Logger   *zap.Logger
	Port     string
	JobLabel string // 対象とする job_label
	DBClient DBClient
}

// 各 job_label ごとにこの metrics を出力
type metrics struct {
	ParsedEntryNum prometheus.GaugeVec
	NewEntryNum    prometheus.GaugeVec
	UpdatedAt      prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		ParsedEntryNum: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "parsed_entry_num",
			Help: "the number of parsed mf records in the job",
		},
			[]string{"job_label"},
		),
		NewEntryNum: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "new_entry_num",
			Help: "the number of NEW parsed mf records in the job",
		},
			[]string{"job_label"},
		),
		UpdatedAt: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "updated_at",
			Help: "the latest time running this job",
		},
			[]string{"job_label"},
		),
	}
	reg.MustRegister(m.ParsedEntryNum)
	reg.MustRegister(m.NewEntryNum)
	reg.MustRegister(m.UpdatedAt)

	return m
}

func (m *MetricsServer) refresh(ctx context.Context, metrics *metrics) {
	const refreshPeriod = 30
	go func() {
		for {
			// 更新処理
			ih, err := m.DBClient.GetLastDetailHistoryWhereJobLabel(ctx, joblabel)
			if err != nil {
				m.Logger.Error("failed to get metrics info", zap.Error(err))
			}
			metrics.ParsedEntryNum.With(m.JobLabel).Set(float64(ih.ParsedEntryNum))
			metrics.NewEntryNum.With(m.JobLabel).Set(float64(ih.NewEntryNum))
			metrics.UpdatedAt.With(m.JobLabel).Set(float64(ih.UpdatedAt.Unix()))

			time.Sleep(refreshPeriod * time.Second)
		}
	}()
	<-ctx.Done()
	m.Logger.Info("refresh routine close")
}

func (m *MetricsServer) Start(ctx context.Context) error {
	m.Logger.Info("metrics server start", zap.String("port", m.Port))
	// Create a non-global registry.
	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	metrics := NewMetrics(reg)
	go m.refresh(ctx, metrics)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", m.Port),
		Handler: nil,
	}
	go func() {
		<-ctx.Done()
		m.Logger.Info("shutdown signal catch")
		ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		nerr := server.Shutdown(ctx2)
		if nerr != nil {
			m.Logger.Error("gracefully shutdown error", zap.Error(nerr))
		}
	}()

	m.Logger.Info("start listening", zap.String("port", m.Port))
	err := server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		// expected error
		err = nil
	} else {
		m.Logger.Error("metrics server close error", zap.Error(err))
		return err
	}

	m.Logger.Info("metrics server shutdown")
	return nil
}
