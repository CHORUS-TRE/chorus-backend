package provider

import (
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/metrics"
)

var serverMetricsOnce sync.Once
var serverMetrics *metrics.ServerMetrics

func ProvideServerMetrics() *metrics.ServerMetrics {
	serverMetricsOnce.Do(func() {
		serverMetrics = metrics.NewServerMetrics()
	})
	return serverMetrics
}
