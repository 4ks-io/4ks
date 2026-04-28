package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// stubProber is a test double for Prober.
type stubProber struct {
	name  string
	err   error
	delay time.Duration
}

func (s *stubProber) Name() string { return s.name }

func (s *stubProber) Probe(ctx context.Context) error {
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return s.err
}

func okProber(name string) *stubProber { return &stubProber{name: name} }
func errProber(name string) *stubProber {
	return &stubProber{name: name, err: errors.New("connection refused")}
}

func newTestRouter(version string, deps SystemControllerDeps) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	c := NewSystemController(version, deps)
	r.GET("/api/ready", c.CheckReadiness)
	r.GET("/api/healthcheck", c.Healthcheck)
	return r
}

func allOKDeps() SystemControllerDeps {
	return SystemControllerDeps{
		DB:        okProber("firestore"),
		Search:    okProber("typesense"),
		Messaging: okProber("pubsub"),
		Storage:   okProber("gcs"),
	}
}

// ── /api/ready ───────────────────────────────────────────────────────────────

func TestCheckReadiness_AlwaysOK(t *testing.T) {
	r := newTestRouter("1.0.0", allOKDeps())

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ready", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCheckReadiness_IndependentOfDependencyHealth(t *testing.T) {
	r := newTestRouter("1.0.0", SystemControllerDeps{
		DB:        errProber("firestore"),
		Search:    errProber("typesense"),
		Messaging: errProber("pubsub"),
		Storage:   errProber("gcs"),
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/ready", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("liveness must be shallow — expected 200 even when deps are down, got %d", rec.Code)
	}
}

// ── /api/healthcheck ─────────────────────────────────────────────────────────

func TestHealthcheck_AllHealthy(t *testing.T) {
	r := newTestRouter("2.0.0", allOKDeps())

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/healthcheck", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body["version"] != "2.0.0" {
		t.Fatalf("expected version 2.0.0, got %v", body["version"])
	}

	db := body["database"].(map[string]interface{})
	if db["provider"] != "firestore" {
		t.Fatalf("expected firestore provider, got %v", db["provider"])
	}
	if db["status"] != "ok" {
		t.Fatalf("expected db status ok, got %v", db["status"])
	}

	storage := body["storage"].(map[string]interface{})
	if storage["provider"] != "gcs" {
		t.Fatalf("expected gcs provider, got %v", storage["provider"])
	}
	if storage["status"] != "ok" {
		t.Fatalf("expected storage status ok, got %v", storage["status"])
	}

	services := body["services"].(map[string]interface{})
	if services["search"] != "ok" {
		t.Fatalf("expected search ok, got %v", services["search"])
	}
	if services["messaging"] != "ok" {
		t.Fatalf("expected messaging ok, got %v", services["messaging"])
	}
}

func TestHealthcheck_DependencyFailed(t *testing.T) {
	deps := allOKDeps()
	deps.DB = errProber("firestore")
	r := newTestRouter("2.0.0", deps)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/healthcheck", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("healthcheck is diagnostic — expected 200 regardless, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	db := body["database"].(map[string]interface{})
	if db["status"] == "ok" {
		t.Fatal("expected db status to contain error, got ok")
	}

	services := body["services"].(map[string]interface{})
	if services["search"] != "ok" {
		t.Fatalf("expected search ok, got %v", services["search"])
	}
}

func TestHealthcheck_Timeout(t *testing.T) {
	deps := allOKDeps()
	deps.Search = &stubProber{name: "typesense", delay: readinessTimeout + 500*time.Millisecond}
	r := newTestRouter("2.0.0", deps)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/healthcheck", nil))

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	services := body["services"].(map[string]interface{})
	if services["search"] == "ok" {
		t.Fatal("expected search to report timeout error, got ok")
	}
}
