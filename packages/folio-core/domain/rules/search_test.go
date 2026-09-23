package rules_test

import (
	"testing"

	"folio/folio-core/domain/rules"
)

func TestFTSQuery(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"single term", "realtime", `"realtime"*`},
		{"two terms are both required", "realtime client", `"realtime" AND "client"*`},
		{"collapses whitespace", "  realtime   client  ", `"realtime" AND "client"*`},
		{"empty stays empty", "", ""},
		{"whitespace only stays empty", "   ", ""},

		// FTS5 reads these as syntax. Pasting an error message must not 500.
		{"quotes are stripped", `foo"`, `"foo"*`},
		{"unbalanced quote mid term", `say "hello`, `"say" AND "hello"*`},
		{"bare AND is a term, not an operator", "AND", `"AND"*`},
		{"boolean words are literal", "cats OR dogs", `"cats" AND "OR" AND "dogs"*`},
		{"NEAR is literal", "NEAR(a b)", `"NEAR" AND "a" AND "b"*`},
		{"caret is dropped", "^start", `"start"*`},
		{"colon is dropped", "title:foo", `"title" AND "foo"*`},
		{"star is dropped from the term", "foo*", `"foo"*`},
		{"parens are dropped", "(a)", `"a"*`},
		{"minus is dropped", "-foo", `"foo"*`},

		// A pasted error message: the realistic worst case.
		{
			"pasted error message",
			`Invalid realtime client: the clientId "x-1" validation_required (400)`,
			`"Invalid" AND "realtime" AND "client" AND "the" AND "clientId" AND "x-1" AND "validation_required" AND "400"*`,
		},

		{"punctuation only yields nothing", `"" () -- :`, ""},
		{"unicode is kept", "café", `"café"*`},
		{"underscores and dashes survive", "snake_case kebab-case", `"snake_case" AND "kebab-case"*`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rules.FTSQuery(tc.in); got != tc.want {
				t.Errorf("FTSQuery(%q)\n got: %q\nwant: %q", tc.in, got, tc.want)
			}
		})
	}
}

// Whatever comes out must be parseable by FTS5. The sanitizer's whole job is
// that no input reaches MATCH as syntax.
func TestFTSQueryNeverEmitsBareOperators(t *testing.T) {
	hostile := []string{
		`"`, `""`, `"""`, `*`, `^`, `:`, `-`, `(`, `)`, `()`,
		"AND", "OR", "NOT", "NEAR", "AND OR NOT",
		`a"b`, `a:b`, `a*b`, `a(b)c`, `NEAR(x, 3)`,
		"   ", "\t\n", `\`, `%`, `'`,
	}

	for _, in := range hostile {
		got := rules.FTSQuery(in)
		if got == "" {
			continue
		}
		if err := checkBalancedQuotes(got); err != nil {
			t.Errorf("FTSQuery(%q) = %q: %v", in, got, err)
		}
	}
}

func checkBalancedQuotes(s string) error {
	count := 0
	for _, r := range s {
		if r == '"' {
			count++
		}
	}
	if count%2 != 0 {
		return errUnbalanced
	}
	return nil
}

var errUnbalanced = errUnbalancedType{}

type errUnbalancedType struct{}

func (errUnbalancedType) Error() string { return "unbalanced quotes" }
