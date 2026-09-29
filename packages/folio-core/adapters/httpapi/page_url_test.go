package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPageURLUsesTheConfiguredBase(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://api.internal/x", nil)
	got := pageURL("https://folio.example/", req, "/acme/web/redesign/tickets/t/interview")
	if want := "https://folio.example/acme/web/redesign/tickets/t/interview"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPageURLFallsBackToTheRequestOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://folio.journ.app/x", nil)
	if got, want := pageURL("", req, "/a/b"), "http://folio.journ.app/a/b"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPageURLTrustsTheForwardedScheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://folio.journ.app/x", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	if got, want := pageURL("", req, "/a/b"), "https://folio.journ.app/a/b"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPageURLIsEmptyWithoutAPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://folio.journ.app/x", nil)
	if got := pageURL("https://folio.example", req, ""); got != "" {
		t.Errorf("a project with no domain has no page, got %q", got)
	}
}
