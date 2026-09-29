package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func stderrOf(t *testing.T, fn func()) string {
	t.Helper()

	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	fn()
	_ = w.Close()
	os.Stderr = stderr

	printed, _ := io.ReadAll(r)
	_ = r.Close()
	return string(printed)
}

func TestNoteIfPagedWarnsOnADefaultPageEvenInJSON(t *testing.T) {
	t.Cleanup(func() { flagJSON = false })

	for _, asJSON := range []bool{false, true} {
		flagJSON = asJSON

		out := stderrOf(t, func() { noteIfPaged(defaultPageSize, 0) })
		if !strings.Contains(out, "default page") {
			t.Errorf("json=%v: stderr = %q, want the default page warning", asJSON, out)
		}
	}
}

func TestNoteIfPagedStaysQuietForADeliberateLimitInJSON(t *testing.T) {
	flagJSON = true
	t.Cleanup(func() { flagJSON = false })

	if out := stderrOf(t, func() { noteIfPaged(20, 20) }); out != "" {
		t.Errorf("stderr = %q, want nothing: the caller chose the limit", out)
	}
}

func TestNoteIfPagedStillWarnsOnALimitOutsideJSON(t *testing.T) {
	flagJSON = false

	if out := stderrOf(t, func() { noteIfPaged(20, 20) }); !strings.Contains(out, "--limit") {
		t.Errorf("stderr = %q, want the limit warning", out)
	}
}

func TestNoteIfPagedIgnoresAShortPage(t *testing.T) {
	flagJSON = false

	if out := stderrOf(t, func() { noteIfPaged(3, 0) }); out != "" {
		t.Errorf("stderr = %q, want nothing", out)
	}
}
