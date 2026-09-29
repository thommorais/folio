package main

import (
	"os"
	"testing"
)

func withStdin(t *testing.T, content string) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.WriteString(content)
	_ = w.Close()

	stdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = stdin
		_ = r.Close()
	})
}

func TestBodyFromRejectsEmptyStdin(t *testing.T) {
	for _, content := range []string{"", "  \n\n"} {
		withStdin(t, content)

		if _, err := bodyFrom("-"); err == nil {
			t.Errorf("bodyFrom(-) with stdin %q: error = nil, want a refusal", content)
		}
	}
}

func TestBodyFromReadsStdin(t *testing.T) {
	withStdin(t, "hello\n")

	got, err := bodyFrom("-")
	if err != nil || got != "hello\n" {
		t.Fatalf("bodyFrom(-) = %q, %v", got, err)
	}
}

func TestBodyFromPassesALiteralThrough(t *testing.T) {
	got, err := bodyFrom("text")
	if err != nil || got != "text" {
		t.Fatalf("bodyFrom(text) = %q, %v", got, err)
	}
	if got, err := bodyFrom(""); err != nil || got != "" {
		t.Fatalf("bodyFrom(empty) = %q, %v, want an unset flag to stay empty", got, err)
	}
}
