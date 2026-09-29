package main

import (
	"strings"
	"testing"
)

func TestProjectRequiredNamesTheChoicesThatSurviveASeparateShell(t *testing.T) {
	t.Setenv("FOLIO_PROJECT", "")
	t.Setenv("FOLIO_CONFIG_DIR", t.TempDir())
	flagProject = ""

	_, err := resolveProject()
	if err == nil {
		t.Fatal("error = nil, want no project")
	}
	for _, want := range []string{"--project", "--here", "eval"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q lacks %q", err, want)
		}
	}
	if strings.Index(err.Error(), "eval") < strings.Index(err.Error(), "--here") {
		t.Errorf("error %q offers eval before --here, but eval does not persist across an agent's shells", err)
	}
}

func TestSelectHintNamesTheChoicesThatSurviveASeparateShell(t *testing.T) {
	hint := selectHint("geral")
	for _, want := range []string{"folio use geral --here", `eval "$(folio use geral)"`} {
		if !strings.Contains(hint, want) {
			t.Errorf("hint %q lacks %q", hint, want)
		}
	}
	if strings.Index(hint, "eval") < strings.Index(hint, "--here") {
		t.Errorf("hint %q offers eval before --here", hint)
	}
}
