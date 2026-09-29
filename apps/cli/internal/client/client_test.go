package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestErrorNamesTheRejectedFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Failed to create record.","data":{"title":{"code":"validation_required","message":"Cannot be blank."},"slug":{"code":"validation_not_unique","message":"Value must be unique."}}}`))
	}))
	defer server.Close()

	_, err := New(server.URL, "tok").GetTicket("x")
	if err == nil {
		t.Fatal("error = nil")
	}

	for _, want := range []string{"400", "Failed to create record.", "slug: Value must be unique.", "title: Cannot be blank."} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q lacks %q", err, want)
		}
	}
}

func TestErrorWithoutFieldsKeepsTheMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not found."}`))
	}))
	defer server.Close()

	_, err := New(server.URL, "tok").GetTicket("x")
	if err == nil || err.Error() != "404: Not found." {
		t.Fatalf("error = %v, want 404: Not found.", err)
	}
}
