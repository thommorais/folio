package main

import (
	"slices"
	"testing"
)

func TestJournalIsNotAliasedToLog(t *testing.T) {
	aliases := journalCommand().Aliases
	for _, taken := range []string{"log", "logs"} {
		if slices.Contains(aliases, taken) {
			t.Errorf("journal aliases %v include %q, which reads as a work log", aliases, taken)
		}
	}
}
