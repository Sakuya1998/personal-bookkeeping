package otel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	api "go.opentelemetry.io/otel/metric"
)

// reset global state between tests
func resetOtel() {
	gin.SetMode(gin.TestMode)
}

// ---------- Init ----------

func TestInit_NilConfig(t *testing.T) {
	resetOtel()
	o, err := Init(nil)
	if err != nil {
		t.Fatalf("Init(nil) unexpected error: %v", err)
	}
	if o == nil {
		t.Fatal("Init(nil) returned nil OTEL")
	}
	if o.Meter != nil {
		t.Error("Init(nil) should not create a Meter")
	}
	// Shutdown must not panic
	o.Shutdown()
}

func TestInit_Disabled(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{Enabled: false})
	if err != nil {
		t.Fatalf("Init(disabled) unexpected error: %v", err)
	}
	if o.Meter != nil {
		t.Error("Init(disabled) should not create a Meter")
	}
	o.Shutdown()
}

func TestInit_EnabledWithPrometheus(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{
		Enabled:        true,
		ServiceName:    "test-service",
		TracesExporter: "none",
		MetricsPath:    "/metrics",
	})
	if err != nil {
		t.Fatalf("Init(enabled) unexpected error: %v", err)
	}
	if o.Meter == nil {
		t.Fatal("Init(enabled) should create a Meter")
	}
	if o.MetricsHandler == nil {
		t.Fatal("Init(enabled) should create a MetricsHandler")
	}
	if o.TracerProvider != nil {
		t.Error("Init with TracesExporter=none should not create TracerProvider")
	}
	o.Shutdown()
}

func TestInit_EnabledWithStdoutTrace(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{
		Enabled:        true,
		ServiceName:    "test-service-trace",
		TracesExporter: "stdout",
		MetricsPath:    "/metrics",
	})
	if err != nil {
		t.Fatalf("Init(stdout-trace) unexpected error: %v", err)
	}
	if o.TracerProvider == nil {
		t.Fatal("Init with TracesExporter=stdout should create TracerProvider")
	}
	o.Shutdown()
}

// ---------- Shutdown ----------

func TestShutdown_NilReceiver(t *testing.T) {
	var o *OTEL
	// must not panic
	o.Shutdown()
}

func TestShutdown_EmptyOTEL(t *testing.T) {
	o := &OTEL{}
	o.Shutdown() // must not panic
}

func TestShutdown_CalledTwice(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{
		Enabled:        true,
		ServiceName:    "test-shutdown-twice",
		TracesExporter: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	o.Shutdown()
	o.Shutdown() // second call must not panic
}

// ---------- GinMiddleware ----------

func TestGinMiddleware_NilMeter(t *testing.T) {
	resetOtel()
	o := &OTEL{} // Meter is nil
	mw := o.GinMiddleware("test")
	if mw == nil {
		t.Fatal("GinMiddleware() returned nil")
	}

	// Test the noop middleware works: it should call c.Next() without panic
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	mw(c)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGinMiddleware_RecordsMetrics(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{
		Enabled:        true,
		ServiceName:    "test-metrics",
		TracesExporter: "none",
		MetricsPath:    "/metrics",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer o.Shutdown()

	mw := o.GinMiddleware("test-metrics")
	if mw == nil {
		t.Fatal("GinMiddleware() returned nil")
	}

	// Simulate a request through the middleware
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/health", nil)
	c.Request = c.Request.WithContext(c.Request.Context())
	mw(c)

	// After one request, the metrics endpoint should show counter > 0
	metricsResp := httptest.NewRecorder()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	o.MetricsHandler.ServeHTTP(metricsResp, metricsReq)

	body := metricsResp.Body.String()
	if len(body) == 0 {
		t.Fatal("metrics endpoint returned empty body")
	}

	// Verify counter metric was recorded
	expectedMetric := "http_server_request_count"
	if !contains(body, expectedMetric) {
		t.Errorf("expected metric %q in response\n%s", expectedMetric, body)
	}
}

// ---------- MetricsHandler HTTP ----------

func TestMetricsHandler_ServesPrometheus(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{
		Enabled:        true,
		ServiceName:    "test-prom",
		TracesExporter: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer o.Shutdown()

	ts := httptest.NewServer(o.MetricsHandler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ---------- Helpers ----------

func TestMetricNames_Present(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{
		Enabled:        true,
		ServiceName:    "test-names",
		TracesExporter: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer o.Shutdown()

	// Verify all built-in metric instruments are created
	tests := []struct {
		name string
		got  interface{}
	}{
		{"reqCounter", o.reqCounter},
		{"durHist", o.durHist},
		{"activeReq", o.activeReq},
	}
	for _, tt := range tests {
		if tt.got == nil {
			t.Errorf("metric %q should not be nil", tt.name)
		}
	}
}

// ---------- Edge cases ----------

func TestInit_InvalidTraceExporter(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{
		Enabled:        true,
		ServiceName:    "test-invalid",
		TracesExporter: "invalid",
	})
	// Invalid trace exporter should not cause an error — we just skip tracing
	if err != nil {
		t.Fatalf("Init(invalid-traces) unexpected error: %v", err)
	}
	if o == nil {
		t.Fatal("Init returned nil")
	}
	if o.TracerProvider != nil {
		t.Error("Init with invalid TracesExporter should skip tracer creation")
	}
	if o.Meter == nil {
		t.Error("Init should still create metrics even with invalid tracer")
	}
	o.Shutdown()
}

// ---------- helpers ----------

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure OTEL handles nil safely
func TestInit_EnabledWithoutMeter(t *testing.T) {
	// Directly test disabled path — Init(nil) must not create meter
	o, _ := Init(nil)
	if o.Meter != nil {
		t.Error("expected nil Meter for disabled OTEL")
	}
}

func TestMeterInterface_Conformance(t *testing.T) {
	resetOtel()
	o, err := Init(&Config{
		Enabled:        true,
		ServiceName:    "test-conformance",
		TracesExporter: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer o.Shutdown()

	// Check that the meter exposes the expected API interface
	if _, ok := interface{}(o.Meter).(api.Meter); !ok {
		t.Error("o.Meter does not implement api.Meter interface")
	}
}
