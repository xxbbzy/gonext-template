package main

import (
	"testing"

	"github.com/xxbbzy/gonext-template/backend/internal/config"
)

func TestNewHTTPMetricsDisabledByDefault(t *testing.T) {
	cfg := &config.Config{}

	if got := newHTTPMetrics(cfg); got != nil {
		t.Fatal("newHTTPMetrics() returned collector when metrics are disabled")
	}
}

func TestNewHTTPMetricsEnabled(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			MetricsEnabled: true,
		},
	}

	if got := newHTTPMetrics(cfg); got == nil {
		t.Fatal("newHTTPMetrics() = nil, want collector when metrics are enabled")
	}
}

func TestNewPrometheusRegistryDisabledWhenMetricsOff(t *testing.T) {
	cfg := &config.Config{}

	registry, err := newPrometheusRegistry(cfg, nil)
	if err != nil {
		t.Fatalf("newPrometheusRegistry() error = %v", err)
	}
	if registry != nil {
		t.Fatal("newPrometheusRegistry() returned registry when metrics are disabled")
	}
}

func TestNewPrometheusRegistryEnabledWhenMetricsOn(t *testing.T) {
	cfg := &config.Config{
		Observability: config.ObservabilityConfig{
			MetricsEnabled: true,
		},
	}

	httpMetrics := newHTTPMetrics(cfg)
	registry, err := newPrometheusRegistry(cfg, httpMetrics)
	if err != nil {
		t.Fatalf("newPrometheusRegistry() error = %v", err)
	}
	if registry == nil {
		t.Fatal("newPrometheusRegistry() = nil, want registry when metrics are enabled")
	}
}
