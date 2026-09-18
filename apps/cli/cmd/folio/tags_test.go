package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"folio/cli/internal/client"
)

func TestParseTagsAcceptsKnownTags(t *testing.T) {
	tags, err := parseTags("frontend,bug")
	if err != nil {
		t.Fatalf("parseTags: %v", err)
	}
	if len(tags) != 2 || tags[0] != "frontend" || tags[1] != "bug" {
		t.Fatalf("got %v, want [frontend bug]", tags)
	}
}

func TestParseTagsNormalizesCaseAndSpacing(t *testing.T) {
	tags, err := parseTags(" Frontend , BUG ")
	if err != nil {
		t.Fatalf("parseTags: %v", err)
	}
	if len(tags) != 2 || tags[0] != "frontend" || tags[1] != "bug" {
		t.Fatalf("got %v, want [frontend bug]", tags)
	}
}

func TestParseTagsSuggestsTheNearestKnownTag(t *testing.T) {
	_, err := parseTags("frontnd")
	if err == nil {
		t.Fatal("want an error for an unknown tag")
	}
	if !strings.Contains(err.Error(), `did you mean "frontend"`) {
		t.Fatalf("got %q, want a frontend suggestion", err)
	}
}

func TestParseTagsReportsEveryUnknownTagAtOnce(t *testing.T) {
	_, err := parseTags("frontnd,bakcend")
	if err == nil {
		t.Fatal("want an error for unknown tags")
	}
	for _, want := range []string{"frontnd", "bakcend"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("got %q, want it to name %q", err, want)
		}
	}
}

// A word far from every known tag gets no suggestion, so the CLI does not bend
// an intentional new tag onto an unrelated one.
func TestParseTagsOffersNoSuggestionForADistantWord(t *testing.T) {
	_, err := parseTags("kubernetes")
	if err == nil {
		t.Fatal("want an error for an unknown tag")
	}
	if strings.Contains(err.Error(), "did you mean") {
		t.Fatalf("got %q, want no suggestion", err)
	}
}

func TestParseTagsIgnoresEmptySegments(t *testing.T) {
	tags, err := parseTags("frontend,,")
	if err != nil {
		t.Fatalf("parseTags: %v", err)
	}
	if len(tags) != 1 || tags[0] != "frontend" {
		t.Fatalf("got %v, want [frontend]", tags)
	}
}

func TestCompleteTagsCompletesAfterTheLastComma(t *testing.T) {
	suggestions, _ := completeTags(nil, nil, "frontend,bu")
	if len(suggestions) != 1 || suggestions[0] != "frontend,bug" {
		t.Fatalf("got %v, want [frontend,bug]", suggestions)
	}
}

func TestCompleteTagsSkipsTagsAlreadyChosen(t *testing.T) {
	suggestions, _ := completeTags(nil, nil, "bug,b")
	for _, suggestion := range suggestions {
		if suggestion == "bug,bug" {
			t.Fatal("suggested a tag already in the list")
		}
	}
}

// A census that silently stops at the server's default page reports tallies
// that are wrong rather than partial, which is the one thing this command
// exists to get right.
func TestTagCensusAsksForTheFullPage(t *testing.T) {
	var queries []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query().Get("limit"))
		_, _ = w.Write([]byte(`{"issues":[],"plans":[],"entries":[]}`))
	}))
	t.Cleanup(server.Close)

	t.Setenv("FOLIO_URL", server.URL)
	t.Setenv("FOLIO_TOKEN", "tok")

	folio, err := api()
	if err != nil {
		t.Fatalf("api: %v", err)
	}
	if _, err := collectTags(folio, "p1"); err != nil {
		t.Fatalf("collectTags: %v", err)
	}

	if len(queries) != 4 {
		t.Fatalf("requests = %d, want 4 (todos, plans, journal, docs)", len(queries))
	}
	want := strconv.Itoa(client.MaxPageSize)
	for i, limit := range queries {
		if limit != want {
			t.Errorf("request %d limit = %q, want %q", i, limit, want)
		}
	}
}
