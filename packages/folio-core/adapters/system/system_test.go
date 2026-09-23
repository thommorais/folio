package system_test

import (
	"regexp"
	"testing"

	"folio/folio-core/adapters/system"
)

// The shares collection rejects any other shape, so a token the service
// generates must match the pattern PocketBase autogenerates.
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
