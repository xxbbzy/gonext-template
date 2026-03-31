package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"

	"github.com/xxbbzy/gonext-template/backend/internal/observability"
	"github.com/xxbbzy/gonext-template/backend/pkg/errcode"
)

func TestMetricsEndpointExposesApplicationAndRuntimeMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router, registry := newMetricsTestRouter(t, true)
	router.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	performRequest(t, router, http.MethodGet, "/ok")

	resp := performRequest(t, router, http.MethodGet, MetricsEndpointPath)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if got := resp.Header().Get("Content-Type"); !strings.Contains(got, "text/plain") {
		t.Fatalf("content-type = %q, want Prometheus text exposition", got)
	}
	if payload := resp.Body.String(); !strings.Contains(payload, "http_requests_total") {
		t.Fatal("metrics payload missing http_requests_total family")
	}
	if payload := resp.Body.String(); !strings.Contains(payload, "http_request_duration_seconds") {
		t.Fatal("metrics payload missing http_request_duration_seconds family")
	}
	if payload := resp.Body.String(); !strings.Contains(payload, "go_goroutines") {
		t.Fatal("metrics payload missing go_goroutines family")
	}
	if payload := resp.Body.String(); !strings.Contains(payload, "process_cpu_seconds_total") {
		t.Fatal("metrics payload missing process_cpu_seconds_total family")
	}

	families := gatherMetricFamilies(t, registry)
	assertMetricFamilyPresent(t, families, "http_requests_total")
	assertMetricFamilyPresent(t, families, "http_request_duration_seconds")
	assertMetricFamilyPresent(t, families, "go_goroutines")
	assertMetricFamilyPresent(t, families, "process_cpu_seconds_total")

	if got := counterValue(t, families, "http_requests_total", map[string]string{
		"method": http.MethodGet,
		"route":  "/ok",
		"status": "200",
	}); got != 1 {
		t.Fatalf("http_requests_total{/ok,200} = %v, want 1", got)
	}
}

func TestHTTPMetricsRecordsSuccessErrorUnmatchedAndRecoveredPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router, registry := newMetricsTestRouter(t, false)
	router.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/app-error", func(c *gin.Context) {
		_ = c.Error(errcode.New(http.StatusUnauthorized, errcode.ErrUnauthorized, "unauthorized"))
	})
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	if code := performRequest(t, router, http.MethodGet, "/ok").Code; code != http.StatusNoContent {
		t.Fatalf("/ok status = %d, want %d", code, http.StatusNoContent)
	}
	if code := performRequest(t, router, http.MethodGet, "/app-error").Code; code != http.StatusUnauthorized {
		t.Fatalf("/app-error status = %d, want %d", code, http.StatusUnauthorized)
	}
	if code := performRequest(t, router, http.MethodGet, "/panic").Code; code != http.StatusInternalServerError {
		t.Fatalf("/panic status = %d, want %d", code, http.StatusInternalServerError)
	}
	if code := performRequest(t, router, http.MethodGet, "/not-found").Code; code != http.StatusNotFound {
		t.Fatalf("/not-found status = %d, want %d", code, http.StatusNotFound)
	}

	resp := performRequest(t, router, http.MethodGet, MetricsEndpointPath)
	if resp.Code != http.StatusOK {
		t.Fatalf("/metrics status = %d, want %d", resp.Code, http.StatusOK)
	}
	families := gatherMetricFamilies(t, registry)

	assertCounterValue(t, families, map[string]string{"method": http.MethodGet, "route": "/ok", "status": "204"}, 1)
	assertCounterValue(t, families, map[string]string{"method": http.MethodGet, "route": "/app-error", "status": "401"}, 1)
	assertCounterValue(t, families, map[string]string{"method": http.MethodGet, "route": "/panic", "status": "500"}, 1)
	assertCounterValue(t, families, map[string]string{"method": http.MethodGet, "route": "unmatched", "status": "404"}, 1)

	if got := histogramCount(t, families, "http_request_duration_seconds", map[string]string{"method": http.MethodGet, "route": "/panic"}); got < 1 {
		t.Fatalf("histogram count for /panic = %v, want >= 1", got)
	}
}

func TestHTTPMetricsExcludeSelfScrapesAndAggregateDynamicRouteLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router, registry := newMetricsTestRouter(t, false)
	router.GET("/items/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	performRequest(t, router, http.MethodGet, "/items/1")
	performRequest(t, router, http.MethodGet, "/items/2")

	firstScrape := performRequest(t, router, http.MethodGet, MetricsEndpointPath)
	if firstScrape.Code != http.StatusOK {
		t.Fatalf("first /metrics status = %d, want %d", firstScrape.Code, http.StatusOK)
	}
	families := gatherMetricFamilies(t, registry)

	assertCounterValue(t, families, map[string]string{"method": http.MethodGet, "route": "/items/:id", "status": "200"}, 2)
	assertCounterSeriesAbsent(t, families, map[string]string{"method": http.MethodGet, "route": "/items/1", "status": "200"})
	assertCounterSeriesAbsent(t, families, map[string]string{"method": http.MethodGet, "route": "/items/2", "status": "200"})
	assertCounterSeriesAbsent(t, families, map[string]string{"method": http.MethodGet, "route": MetricsEndpointPath, "status": "200"})

	secondScrape := performRequest(t, router, http.MethodGet, MetricsEndpointPath)
	if secondScrape.Code != http.StatusOK {
		t.Fatalf("second /metrics status = %d, want %d", secondScrape.Code, http.StatusOK)
	}
	families = gatherMetricFamilies(t, registry)
	assertCounterValue(t, families, map[string]string{"method": http.MethodGet, "route": "/items/:id", "status": "200"}, 2)
}

func newMetricsTestRouter(t *testing.T, includeRuntimeCollectors bool) (*gin.Engine, *prometheus.Registry) {
	t.Helper()

	httpMetrics := NewHTTPMetrics(HTTPMetricsOptions{})
	registry, err := observability.NewPrometheusRegistry(observability.RegistryOptions{
		IncludeRuntimeCollectors: includeRuntimeCollectors,
		ApplicationCollectors:    httpMetrics.Collectors(),
	})
	if err != nil {
		t.Fatalf("new prometheus registry: %v", err)
	}

	router := gin.New()
	router.Use(
		RequestID(),
		httpMetrics.Middleware(),
		Recovery(zap.NewNop()),
		RequestLogger(zap.NewNop()),
		ErrorHandler(),
	)
	router.GET(MetricsEndpointPath, gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	return router, registry
}

func performRequest(t *testing.T, router *gin.Engine, method string, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func gatherMetricFamilies(t *testing.T, registry *prometheus.Registry) map[string]*dto.MetricFamily {
	t.Helper()

	gathered, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	families := make(map[string]*dto.MetricFamily, len(gathered))
	for _, family := range gathered {
		families[family.GetName()] = family
	}
	return families
}

func assertMetricFamilyPresent(t *testing.T, families map[string]*dto.MetricFamily, family string) {
	t.Helper()

	if _, ok := families[family]; !ok {
		t.Fatalf("metric family %q not found", family)
	}
}

func assertCounterValue(t *testing.T, families map[string]*dto.MetricFamily, labels map[string]string, want float64) {
	t.Helper()

	if got := counterValue(t, families, "http_requests_total", labels); got != want {
		t.Fatalf("http_requests_total%v = %v, want %v", labels, got, want)
	}
}

func assertCounterSeriesAbsent(t *testing.T, families map[string]*dto.MetricFamily, labels map[string]string) {
	t.Helper()

	if _, found := findCounterMetric(families, "http_requests_total", labels); found {
		t.Fatalf("unexpected http_requests_total series found for labels %v", labels)
	}
}

func counterValue(t *testing.T, families map[string]*dto.MetricFamily, family string, labels map[string]string) float64 {
	t.Helper()

	metric, found := findCounterMetric(families, family, labels)
	if !found {
		t.Fatalf("counter series not found for %s with labels %v", family, labels)
	}
	return metric.GetCounter().GetValue()
}

func histogramCount(t *testing.T, families map[string]*dto.MetricFamily, family string, labels map[string]string) float64 {
	t.Helper()

	metricFamily, ok := families[family]
	if !ok {
		t.Fatalf("metric family %q not found", family)
	}

	for _, metric := range metricFamily.GetMetric() {
		if metric.GetHistogram() == nil {
			continue
		}
		if labelsMatch(metric.GetLabel(), labels) {
			return float64(metric.GetHistogram().GetSampleCount())
		}
	}

	t.Fatalf("histogram series not found for %s with labels %v", family, labels)
	return 0
}

func findCounterMetric(families map[string]*dto.MetricFamily, family string, labels map[string]string) (*dto.Metric, bool) {
	metricFamily, ok := families[family]
	if !ok {
		return nil, false
	}

	for _, metric := range metricFamily.GetMetric() {
		if metric.GetCounter() == nil {
			continue
		}
		if labelsMatch(metric.GetLabel(), labels) {
			return metric, true
		}
	}
	return nil, false
}

func labelsMatch(pairs []*dto.LabelPair, want map[string]string) bool {
	if len(want) == 0 {
		return len(pairs) == 0
	}

	got := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		got[pair.GetName()] = pair.GetValue()
	}
	if len(got) != len(want) {
		return false
	}

	for key, value := range want {
		if got[key] != value {
			return false
		}
	}

	return true
}
