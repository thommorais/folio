package rules_test

import (
	"errors"
	"strings"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func TestValidateProject(t *testing.T) {
	t.Run("rejects an empty name", func(t *testing.T) {
		err := rules.ValidateProject(domain.Project{Slug: "api", Name: "  "})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("rejects a name over the limit", func(t *testing.T) {
		err := rules.ValidateProject(domain.Project{Slug: "api", Name: strings.Repeat("a", rules.NameMaxLen+1)})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("rejects a slug with invalid characters", func(t *testing.T) {
		for _, slug := range []string{"Has Space", "UPPER", "trailing-", "-leading", "sym@bol", ""} {
			if err := rules.ValidateProject(domain.Project{Slug: slug, Name: "ok"}); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("slug %q: want validation error, got %v", slug, err)
			}
		}
	})

	t.Run("accepts a kebab-case slug", func(t *testing.T) {
		for _, slug := range []string{"api", "my-api", "api2", "a"} {
			if err := rules.ValidateProject(domain.Project{Slug: slug, Name: "ok"}); err != nil {
				t.Errorf("slug %q: want nil, got %v", slug, err)
			}
		}
	})
}

func TestValidatePlan(t *testing.T) {
	valid := domain.Plan{ProjectID: "p1", Title: "Ship search", Status: domain.PlanActive}

	t.Run("accepts a valid plan", func(t *testing.T) {
		if err := rules.ValidatePlan(valid); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("requires a project", func(t *testing.T) {
		p := valid
		p.ProjectID = ""
		if !errors.Is(rules.ValidatePlan(p), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("requires a title", func(t *testing.T) {
		p := valid
		p.Title = ""
		if !errors.Is(rules.ValidatePlan(p), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("rejects an unknown status", func(t *testing.T) {
		p := valid
		p.Status = "sideways"
		if !errors.Is(rules.ValidatePlan(p), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})
}

func TestProgressPercentEmpty(t *testing.T) {
	if (domain.Progress{}).Percent() != 0 {
		t.Fatal("an empty plan must report 0 percent, not divide by zero")
	}
}

func TestValidateJournalEntry(t *testing.T) {
	valid := domain.JournalEntry{ProjectID: "p1", Slug: "restored-ga4-pageview-tracking", Title: "Restored GA4 pageview tracking", Body: "The config call was deleted."}

	if err := rules.ValidateJournalEntry(valid); err != nil {
		t.Fatalf("want nil, got %v", err)
	}

	t.Run("requires a title", func(t *testing.T) {
		bad := valid
		bad.Title = "   "
		if !errors.Is(rules.ValidateJournalEntry(bad), domain.ErrValidation) {
			t.Fatal("an entry without a title could never be found again")
		}
	})

	t.Run("requires a slug", func(t *testing.T) {
		bad := valid
		bad.Slug = ""
		if !errors.Is(rules.ValidateJournalEntry(bad), domain.ErrValidation) {
			t.Fatal("an entry without a slug has no URL")
		}
	})

	t.Run("rejects a slug that is not kebab-case", func(t *testing.T) {
		bad := valid
		bad.Slug = "Restored GA4"
		if !errors.Is(rules.ValidateJournalEntry(bad), domain.ErrValidation) {
			t.Fatal("a slug with spaces or capitals would not survive a URL")
		}
	})

	t.Run("allows an empty body", func(t *testing.T) {
		stub := valid
		stub.Body = ""
		if err := rules.ValidateJournalEntry(stub); err != nil {
			t.Fatalf("work is often logged before it is finished, got %v", err)
		}
	})

	t.Run("requires a project", func(t *testing.T) {
		bad := valid
		bad.ProjectID = ""
		if !errors.Is(rules.ValidateJournalEntry(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("bounds the work refs", func(t *testing.T) {
		bad := valid
		bad.Branch = strings.Repeat("b", rules.RefMaxLen+1)
		if !errors.Is(rules.ValidateJournalEntry(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})
}

func TestValidateDoc(t *testing.T) {
	valid := domain.Doc{ProjectID: "p1", Slug: "architecture", Title: "Architecture", Body: "hexagonal"}

	if err := rules.ValidateDoc(valid); err != nil {
		t.Fatalf("want nil, got %v", err)
	}

	t.Run("rejects a bad slug", func(t *testing.T) {
		bad := valid
		bad.Slug = "Not A Slug"
		if !errors.Is(rules.ValidateDoc(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})
}

func TestCanRemoveMember(t *testing.T) {
	project := domain.Project{Members: []domain.Member{
		{UserID: "u1", Role: domain.RoleOwner},
		{UserID: "u2", Role: domain.RoleEditor},
	}}

	t.Run("refuses to remove the last owner", func(t *testing.T) {
		if err := rules.CanRemoveMember(project, "u1"); !errors.Is(err, domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("allows removing a non-owner", func(t *testing.T) {
		if err := rules.CanRemoveMember(project, "u2"); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("allows removing an owner when another remains", func(t *testing.T) {
		two := domain.Project{Members: []domain.Member{
			{UserID: "u1", Role: domain.RoleOwner},
			{UserID: "u3", Role: domain.RoleOwner},
		}}
		if err := rules.CanRemoveMember(two, "u1"); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}

func TestSnippet(t *testing.T) {
	t.Run("returns short text unchanged", func(t *testing.T) {
		if got := rules.Snippet("short", 20); got != "short" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("truncates on a word boundary with an ellipsis", func(t *testing.T) {
		got := rules.Snippet("the quick brown fox jumps", 12)
		if !strings.HasSuffix(got, "…") {
			t.Fatalf("want an ellipsis suffix, got %q", got)
		}
		if len([]rune(got)) > 13 {
			t.Fatalf("too long: %q", got)
		}
	})
}
