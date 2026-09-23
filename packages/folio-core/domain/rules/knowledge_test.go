package rules_test

import (
	"strings"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func validKnowledge() domain.Knowledge {
	return domain.Knowledge{
		Slug:      "pocketbase-realtime-on-railway",
		Title:     "Enable realtime in PocketBase on Railway",
		Body:      "Railway buffers the SSE response.",
		CreatedBy: "u-1",
	}
}

func TestValidateKnowledge(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*domain.Knowledge)
		field  string
	}{
		{"valid", func(*domain.Knowledge) {}, ""},

		// A knowledge note belongs to no project by default, which is the
		// whole point: it outlives the work it was learned on.
		{"no project is fine", func(k *domain.Knowledge) { k.ProjectID = "" }, ""},
		{"a project is fine", func(k *domain.Knowledge) { k.ProjectID = "p-1" }, ""},
		{"no body is fine", func(k *domain.Knowledge) { k.Body = "" }, ""},

		{"title is required", func(k *domain.Knowledge) { k.Title = "" }, "title"},
		{"blank title is required", func(k *domain.Knowledge) { k.Title = "   " }, "title"},
		{"title is bounded", func(k *domain.Knowledge) { k.Title = strings.Repeat("x", rules.TitleMaxLen+1) }, "title"},
		{"slug is required", func(k *domain.Knowledge) { k.Slug = "" }, "slug"},
		{"slug is kebab-case", func(k *domain.Knowledge) { k.Slug = "Not A Slug" }, "slug"},
		{"body is bounded", func(k *domain.Knowledge) { k.Body = strings.Repeat("x", rules.BodyMaxLen+1) }, "body"},

		// Knowledge is readable by everyone, so the author is the only record
		// of where it came from and cannot be dropped.
		{"author is required", func(k *domain.Knowledge) { k.CreatedBy = "" }, "created_by"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k := validKnowledge()
			tc.mutate(&k)

			err := rules.ValidateKnowledge(k)
			if tc.field == "" {
				if err != nil {
					t.Fatalf("want valid, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("want %s to be rejected", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error %q does not name %q", err, tc.field)
			}
		})
	}
}
