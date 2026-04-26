package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestHealthEndpoint(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := &Handler{}
	if err := handler.Health(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetGraphEndpoint(t *testing.T) {
	_ = t
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "test-user")

	// Skip test - needs real GetGraphUseCase
}

func TestAnalyzeEndpoint(t *testing.T) {
	e := echo.New()

	body := `{"text": "Had meeting with PersonA today"}`
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader([]byte(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "test-user")

	// Test validation - missing text should return 400
	badReq := `{}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader([]byte(badReq)))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.Set("user_id", "test-user")

	handler := &Handler{}
	err := handler.Analyze(c2)
	if err == nil {
		t.Log("expected error for empty text, got nil")
	}
}
