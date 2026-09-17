package rules

import (
	"strings"
)

// Snippet trims text to max runes for a search result, cutting on a word
// boundary where one is close enough to the limit to be worth keeping.
func Snippet(text string, max int) string {
	clean := strings.Join(strings.Fields(text), " ")
	runes := []rune(clean)
	if len(runes) <= max {
		return clean
	}
	cut := string(runes[:max])
	if idx := strings.LastIndex(cut, " "); idx > max/2 {
		cut = cut[:idx]
	}
	return strings.TrimRight(cut, " ,.;:") + "…"
}
