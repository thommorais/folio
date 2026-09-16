package main

import (
	"fmt"
	"sort"
	"strings"
)

// The known tags are the CLI's opinion, not the API's: folio-core stores
// whatever it is handed. Keeping the list here means a typo is caught at the
// point it is typed, where the fix is one keystroke, rather than becoming a
// tag of its own that nothing will ever query again.
//
// Two axes. Context says where the work lives, so `--tags frontend` narrows a
// whole project to one surface. Kind says what sort of work it is. A todo
// usually carries one of each.
var (
	contextTags = []string{
		"api",
		"backend",
		"cli",
		"db",
		"design",
		"docs",
		"frontend",
		"infra",
		"mcp",
		"mobile",
		"tui",
		"web",
	}

	kindTags = []string{
		"bug",
		"chore",
		"decision",
		"deploy",
		"dx",
		"perf",
		"refactor",
		"release",
		"security",
		"spike",
		"test",
	}
)

func knownTags() []string {
	all := make([]string, 0, len(contextTags)+len(kindTags))
	all = append(all, contextTags...)
	all = append(all, kindTags...)
	sort.Strings(all)
	return all
}

func isKnownTag(tag string) bool {
	for _, known := range knownTags() {
		if known == tag {
			return true
		}
	}
	return false
}

// tagHelp is appended to every --tags flag description so the vocabulary is
// visible at the moment someone reaches for the flag, not only in the long help.
func tagHelp() string {
	return "comma separated tags; context: " + strings.Join(contextTags, ", ") +
		"; kind: " + strings.Join(kindTags, ", ")
}

// parseTags splits the flag and vets every tag, returning the unknown ones with
// a suggestion each so one pass reports every mistake instead of stopping at
// the first.
func parseTags(value string) ([]string, error) {
	raw := strings.Split(value, ",")
	tags := make([]string, 0, len(raw))
	var rejected []string

	for _, entry := range raw {
		tag := strings.ToLower(strings.TrimSpace(entry))
		if tag == "" {
			continue
		}
		if !isKnownTag(tag) {
			rejected = append(rejected, describeUnknown(tag))
			continue
		}
		tags = append(tags, tag)
	}

	if len(rejected) > 0 {
		return nil, fmt.Errorf(
			"unknown %s: %s\nknown tags: %s",
			plural("tag", len(rejected)),
			strings.Join(rejected, "; "),
			strings.Join(knownTags(), ", "),
		)
	}
	return tags, nil
}

func describeUnknown(tag string) string {
	if suggestion, ok := nearest(tag); ok {
		return fmt.Sprintf("%q (did you mean %q?)", tag, suggestion)
	}
	return fmt.Sprintf("%q", tag)
}

// nearest finds the closest known tag within a distance that scales with the
// word, so "frontnd" resolves but "search" is left alone rather than being
// bent onto an unrelated tag.
func nearest(tag string) (string, bool) {
	best := ""
	bestDistance := len(tag)/3 + 1

	for _, known := range knownTags() {
		if d := distance(tag, known); d <= bestDistance {
			best, bestDistance = known, d
		}
	}
	return best, best != ""
}

func distance(a, b string) int {
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)

	for j := range previous {
		previous[j] = j
	}

	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			current[j] = min(previous[j]+1, current[j-1]+1, previous[j-1]+cost)
		}
		copy(previous, current)
	}
	return previous[len(b)]
}

func plural(word string, n int) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
