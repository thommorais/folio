package rules

import (
	"strings"
	"unicode"
)

// FTSQuery turns free text into an FTS5 MATCH expression.
//
// Every term is quoted and the terms are joined with AND, so nothing in the
// input is read as FTS5 syntax. That is the point: an agent pastes an error
// message containing quotes, colons and parentheses, and it has to come back
// as results rather than a parse error. The cost is that operators callers
// might have wanted (OR, NEAR, prefix globs) are literal terms instead.
//
// The final term carries a * so a query still matches while it is being typed,
// which is what makes the web app's search-as-you-type feel live.
//
// An empty result means "no text constraint", not "match nothing" — the caller
// decides what an all-punctuation query does.
func FTSQuery(text string) string {
	terms := strings.FieldsFunc(text, func(r rune) bool {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
		// Kept because they appear inside identifiers an agent searches for,
		// such as snake_case names, kebab-case slugs and hyphenated ids.
		return r != '_' && r != '-'
	})

	quoted := make([]string, 0, len(terms))
	for _, term := range terms {
		// FieldsFunc already dropped every quote, so nothing here can close
		// the string early.
		trimmed := strings.Trim(term, "-_")
		if trimmed == "" {
			continue
		}
		quoted = append(quoted, `"`+trimmed+`"`)
	}

	if len(quoted) == 0 {
		return ""
	}

	return strings.Join(quoted, " AND ") + "*"
}
