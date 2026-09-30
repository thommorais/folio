package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"folio/cli/internal/client"
	"folio/cli/internal/config"
)

func apiFailure(t *testing.T, status int, body string) error {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	_, err := client.New(server.URL, "tok").GetTicket("x")
	if err == nil {
		t.Fatal("error = nil")
	}
	return err
}

func TestExitCodeSaysWhatKindOfFailure(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"unauthorized", apiFailure(t, 401, `{"message":"expired"}`), exitAuth},
		{"forbidden", apiFailure(t, 403, `{"message":"no"}`), exitAuth},
		{"no token", fmt.Errorf("wrapped: %w", config.ErrNoToken), exitAuth},
		{"no client token", client.ErrNoToken, exitAuth},
		{"not found", apiFailure(t, 404, `{"message":"Not found."}`), exitNotFound},
		{"validation", apiFailure(t, 400, `{"message":"Failed."}`), exitValidation},
		{"server error", apiFailure(t, 500, `{"message":"boom"}`), exitFailure},
		{"not an api error", errors.New("nothing to update"), exitFailure},
	}

	for _, tc := range cases {
		if got := exitCode(tc.err); got != tc.want {
			t.Errorf("%s: exitCode = %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestJSONErrorCarriesStatusAndFields(t *testing.T) {
	err := apiFailure(t, 400, `{"message":"Failed to create record.","data":{"title":{"message":"Cannot be blank."}}}`)

	var got struct {
		Error  string            `json:"error"`
		Status int               `json:"status"`
		Fields map[string]string `json:"fields"`
	}
	if err := json.Unmarshal([]byte(errorJSON(err)), &got); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if got.Status != 400 || got.Fields["title"] != "Cannot be blank." || got.Error == "" {
		t.Errorf("got %+v", got)
	}
}

func TestJSONErrorOmitsWhatAPlainErrorLacks(t *testing.T) {
	var got map[string]any
	if err := json.Unmarshal([]byte(errorJSON(errors.New("nothing to update"))), &got); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if got["error"] != "nothing to update" {
		t.Errorf("error = %v", got["error"])
	}
	if _, ok := got["status"]; ok {
		t.Errorf("status present on a non-API error: %v", got)
	}
	if _, ok := got["fields"]; ok {
		t.Errorf("fields present on a non-API error: %v", got)
	}
}
