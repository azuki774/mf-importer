package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	Logger    *zap.Logger
	Port      string
	JobLabels []string // 対象とする job_label のリスト
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

func (m *MetricsServer) Start(ctx context.Context) error {
	m.Logger.Info("metrics server start", zap.String("port", m.Port))
	// Create a non-global registry.
	reg := prometheus.NewRegistry()

	// Create new metrics and register them using the custom registry.
	_ = NewMetrics(reg) // TODO

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

	m.Logger.Info("start listening", zap.String("port", s.Port))
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
