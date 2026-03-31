package observability

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestNewPrometheusRegistryIncludesRuntimeCollectorsWhenEnabled(t *testing.T) {
	custom := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "custom_metric_total",
		Help: "A custom metric for testing.",
	})

	registry, err := NewPrometheusRegistry(RegistryOptions{
		IncludeRuntimeCollectors: true,
		ApplicationCollectors:    []prometheus.Collector{custom},
	})
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	assertFamilyPresent(t, families, "custom_metric_total")
	assertFamilyPresent(t, families, "go_goroutines")
	assertFamilyPresent(t, families, "process_cpu_seconds_total")
}

func TestNewPrometheusRegistrySkipsRuntimeCollectorsWhenDisabled(t *testing.T) {
	custom := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "custom_metric_gauge",
		Help: "A custom metric for testing.",
	})

	registry, err := NewPrometheusRegistry(RegistryOptions{
		IncludeRuntimeCollectors: false,
		ApplicationCollectors:    []prometheus.Collector{custom, nil},
	})
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	assertFamilyPresent(t, families, "custom_metric_gauge")
	assertFamilyAbsent(t, families, "go_goroutines")
	assertFamilyAbsent(t, families, "process_cpu_seconds_total")
}

func assertFamilyPresent(t *testing.T, families []*dto.MetricFamily, family string) {
	t.Helper()

	for _, metricFamily := range families {
		if metricFamily.GetName() == family {
			return
		}
	}
	t.Fatalf("metric family %q not found", family)
}

func assertFamilyAbsent(t *testing.T, families []*dto.MetricFamily, family string) {
	t.Helper()

	for _, metricFamily := range families {
		if metricFamily.GetName() == family {
			t.Fatalf("metric family %q unexpectedly found", family)
		}
	}
}
