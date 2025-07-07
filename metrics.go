package eventServer

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "event_server"
)

type statsExporter struct {
	events *uint64

	eventsDesc *prometheus.Desc
}

func (s *Plugin) MetricsCollector() []prometheus.Collector {
	// p - implements Exporter interface (workers)
	return []prometheus.Collector{s.metrics}
}

func (se *statsExporter) CountEvents() {
	atomic.AddUint64(se.events, 1)
}

func newStatsExporter() *statsExporter {
	return &statsExporter{
		events: toPtr(uint64(0)),

		eventsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "events"), "Number of events registered in the directory", nil, nil),
	}
}

func (se *statsExporter) Describe(d chan<- *prometheus.Desc) {
	// send description
	d <- se.eventsDesc
}

func (se *statsExporter) Collect(ch chan<- prometheus.Metric) {
	// send the values to the prometheus
	ch <- prometheus.MustNewConstMetric(se.eventsDesc, prometheus.GaugeValue, float64(atomic.LoadUint64(se.events)))
}

func toPtr[T any](v T) *T {
	return &v
}
