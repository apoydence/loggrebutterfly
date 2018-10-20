package rangemetrics

import (
	"github.com/poy/loggrebutterfly/master/internal/rangemetrics/networkreader"
	"github.com/poy/petasos/metrics"
	"github.com/poy/petasos/router"
)

type RangeMetrics struct {
	delta *metrics.Delta
}

func New(addrs []string) *RangeMetrics {
	return &RangeMetrics{
		delta: metrics.NewDelta(10000, metrics.NewReader(addrs, networkreader.New())),
	}
}

func (m *RangeMetrics) Metrics(file string) (metric router.Metric, err error) {
	return m.delta.Metrics(file)
}
