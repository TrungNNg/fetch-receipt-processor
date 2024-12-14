package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRecoverPanic(t *testing.T) {

	app := newTestApplication()

	// Initialize a new httptest.ResponseRecorder and dummy http.Request.
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP handler that panic
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	app.recoverPanic(next).ServeHTTP(rr, r)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "the server encountered a problem") {
		t.Errorf("unexpected response body: %s", rr.Body.String())
	}

	if conn := rr.Header().Get("Connection"); conn != "close" {
		t.Errorf("expected Connection header to be 'close', got %s", conn)
	}
}

func TestRateLimit(t *testing.T) {
	app := newTestApplication()

	// Configure rate limiter settings.
	app.config.limiter.enabled = true
	app.config.limiter.rps = 1   // 1 request per second.
	app.config.limiter.burst = 2 // Allow a burst of 2 requests.

	// Define a handler to test the rate limit middleware.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the handler with the rateLimit middleware.
	rateLimitMiddleware := app.rateLimit(testHandler)

	// Simulate requests from the same client IP.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	// Create a response recorder.
	rec := httptest.NewRecorder()

	// First request: should pass.
	rateLimitMiddleware.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Second request: should pass (burst capacity).
	rec = httptest.NewRecorder()
	rateLimitMiddleware.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Third request: should fail (exceeded rate limit).
	rec = httptest.NewRecorder()
	rateLimitMiddleware.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}

	// Wait for rate limiter to replenish.
	time.Sleep(1 * time.Second)

	// Fourth request: should pass again.
	rec = httptest.NewRecorder()
	rateLimitMiddleware.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
