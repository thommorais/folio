package system_test

import (
	"regexp"
	"testing"

	"folio/folio-core/adapters/system"
)

func TestTokenMatchesTheStoredPattern(t *testing.T) {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9]{43}$`)
	seen := map[string]bool{}
	for range 1000 {
		tok := system.TokenGenerator{}.NewToken()
		if !pattern.MatchString(tok) {
			t.Fatalf("token %q does not match %s", tok, pattern)
		}
		if seen[tok] {
			t.Fatalf("token %q repeated", tok)
		}
		seen[tok] = true
	}
}
