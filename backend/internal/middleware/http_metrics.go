package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// MetricsEndpointPath is the operational scrape endpoint.
	MetricsEndpointPath = "/metrics"

	defaultUnmatchedRouteLabel = "unmatched"
)

// HTTPMetricsOptions customizes HTTP metrics middleware behavior.
type HTTPMetricsOptions struct {
	ExcludedPaths          []string
	UnmatchedRouteLabel    string
	RequestDurationBuckets []float64
}

// HTTPMetrics contains Prometheus metric families for backend HTTP traffic.
type HTTPMetrics struct {
	httpRequestsTotal      *prometheus.CounterVec
	httpRequestDurationSec *prometheus.HistogramVec
	excludedPaths          map[string]struct{}
	unmatchedRouteLabel    string
}

// NewHTTPMetrics builds backend HTTP metric collectors.
func NewHTTPMetrics(opts HTTPMetricsOptions) *HTTPMetrics {
	unmatchedRouteLabel := strings.TrimSpace(opts.UnmatchedRouteLabel)
	if unmatchedRouteLabel == "" {
		unmatchedRouteLabel = defaultUnmatchedRouteLabel
	}

	excludedPaths := make(map[string]struct{}, len(opts.ExcludedPaths)+1)
	excludedPaths[MetricsEndpointPath] = struct{}{}
	for _, path := range opts.ExcludedPaths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		excludedPaths[path] = struct{}{}
	}

	return &HTTPMetrics{
		httpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of completed HTTP requests.",
		}, []string{"method", "route", "status"}),
		httpRequestDurationSec: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of completed HTTP requests in seconds.",
			Buckets: opts.RequestDurationBuckets,
		}, []string{"method", "route"}),
		excludedPaths:       excludedPaths,
		unmatchedRouteLabel: unmatchedRouteLabel,
	}
}

// Collectors returns all application collector families owned by this middleware.
func (m *HTTPMetrics) Collectors() []prometheus.Collector {
	if m == nil {
		return nil
	}
	return []prometheus.Collector{m.httpRequestsTotal, m.httpRequestDurationSec}
}

// Middleware returns Gin middleware that records request count/status and duration.
func (m *HTTPMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m == nil || c == nil || c.Request == nil || c.Request.URL == nil {
			c.Next()
			return
		}

		requestPath := c.Request.URL.Path
		if m.isExcludedPath(requestPath) {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		routeLabel := m.routeLabel(c)
		if m.isExcludedPath(routeLabel) {
			return
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())
		m.httpRequestsTotal.WithLabelValues(method, routeLabel, status).Inc()
		m.httpRequestDurationSec.WithLabelValues(method, routeLabel).Observe(time.Since(start).Seconds())
	}
}

func (m *HTTPMetrics) routeLabel(c *gin.Context) string {
	route := strings.TrimSpace(c.FullPath())
	if route == "" {
		return m.unmatchedRouteLabel
	}
	return route
}

func (m *HTTPMetrics) isExcludedPath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	_, excluded := m.excludedPaths[path]
	return excluded
}
