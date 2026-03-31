package observability

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	collectors "github.com/prometheus/client_golang/prometheus/collectors"
)

// RegistryOptions controls how a Prometheus registry is assembled.
type RegistryOptions struct {
	IncludeRuntimeCollectors bool
	ApplicationCollectors    []prometheus.Collector
}

// NewPrometheusRegistry builds a Prometheus registry with optional runtime collectors.
func NewPrometheusRegistry(opts RegistryOptions) (*prometheus.Registry, error) {
	registry := prometheus.NewRegistry()

	if opts.IncludeRuntimeCollectors {
		if err := registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
			return nil, fmt.Errorf("register process collector: %w", err)
		}
		if err := registry.Register(collectors.NewGoCollector()); err != nil {
			return nil, fmt.Errorf("register go collector: %w", err)
		}
	}

	for _, collector := range opts.ApplicationCollectors {
		if collector == nil {
			continue
		}
		if err := registry.Register(collector); err != nil {
			return nil, fmt.Errorf("register application collector: %w", err)
		}
	}

	return registry, nil
}
