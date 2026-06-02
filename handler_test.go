package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestDebugID(t *testing.T) {
	id1 := debugID("192.168.1.1")
	id2 := debugID("192.168.1.1")
	id3 := debugID("10.0.0.1")

	if id1 != id2 {
		t.Error("same IP should produce same debug_id")
	}
	if id1 == id3 {
		t.Error("different IPs should produce different debug_id")
	}
	if len(id1) != debugIDLen {
		t.Errorf("expected debug_id length %d, got %d", debugIDLen, len(id1))
	}
}

func TestTurnoutHandler(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/debug/turnout", TurnoutHandler)
	req := httptest.NewRequest("GET", "/api/debug/turnout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
