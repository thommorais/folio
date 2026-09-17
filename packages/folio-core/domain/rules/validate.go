// Package rules holds the pure business rules of folio: validation, derived
// fields and transition legality. Nothing here touches storage, time zones or
// transport, so every rule is testable in isolation.
package rules

import (
	"regexp"
	"strings"

	"folio/folio-core/domain"
)

// Length limits. Titles stay short because agents list them densely; bodies
// are generous because docs and log meta carry real content.
const (
	NameMaxLen  = 120
	SlugMaxLen  = 60
	TitleMaxLen = 200
	DescrMaxLen = 2000
	BodyMaxLen  = 500000
	// RefMaxLen bounds a branch name, PR number or ticket key.
	RefMaxLen = 200
)

// slugPattern is lowercase kebab-case: the identifier a human types in a CLI
// and a code agent can guess from a project name.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidateSlug checks a URL-safe kebab-case identifier.
func ValidateSlug(field, slug string) error {
	if slug == "" {
		return domain.Invalid(field, "is required")
	}
	if len(slug) > SlugMaxLen {
		return domain.Invalid(field, "must be at most 60 characters")
	}
	if !slugPattern.MatchString(slug) {
		return domain.Invalid(field, "must be lowercase kebab-case, e.g. my-project")
	}
	return nil
}

func required(field, value string, max int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return domain.Invalid(field, "is required")
	}
	if len([]rune(trimmed)) > max {
		return domain.Invalid(field, "is too long")
	}
	return nil
}

func optional(field, value string, max int) error {
	if len([]rune(value)) > max {
		return domain.Invalid(field, "is too long")
	}
	return nil
}

func ValidateProject(p domain.Project) error {
	if err := ValidateSlug("slug", p.Slug); err != nil {
		return err
	}
	if err := required("name", p.Name, NameMaxLen); err != nil {
		return err
	}
	return optional("description", p.Descr, DescrMaxLen)
}

var planStatuses = map[domain.PlanStatus]bool{
	domain.PlanDraft: true, domain.PlanActive: true,
	domain.PlanDone: true, domain.PlanAbandoned: true,
}

func ValidatePlan(p domain.Plan) error {
	if p.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if err := required("title", p.Title, TitleMaxLen); err != nil {
		return err
	}
	if err := optional("goal", p.Goal, DescrMaxLen); err != nil {
		return err
	}
	if !planStatuses[p.Status] {
		return domain.Invalid("status", "must be one of draft, active, done, abandoned")
	}
	return nil
}

var priorities = map[domain.Priority]bool{
	domain.PriorityLow: true, domain.PriorityMedium: true, domain.PriorityHigh: true,
}

// ValidateJournalEntry checks a work log. The body is optional: an entry may be
// created as a stub and filled in as the work proceeds.
func ValidateJournalEntry(e domain.JournalEntry) error {
	if e.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if err := ValidateSlug("slug", e.Slug); err != nil {
		return err
	}
	if err := required("title", e.Title, TitleMaxLen); err != nil {
		return err
	}
	if err := optional("body", e.Body, BodyMaxLen); err != nil {
		return err
	}
	if err := optional("branch", e.Branch, RefMaxLen); err != nil {
		return err
	}
	if err := optional("pr", e.PR, RefMaxLen); err != nil {
		return err
	}
	return optional("external_ref", e.ExternalRef, RefMaxLen)
}

var wayfinderTypes = map[domain.WayfinderType]bool{
	domain.WayfinderMap: true, domain.WayfinderResearch: true, domain.WayfinderPrototype: true,
	domain.WayfinderGrilling: true, domain.WayfinderTask: true,
}

func ValidateDoc(d domain.Doc) error {
	if d.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if err := ValidateSlug("slug", d.Slug); err != nil {
		return err
	}
	if err := required("title", d.Title, TitleMaxLen); err != nil {
		return err
	}
	return optional("body", d.Body, BodyMaxLen)
}

var roles = map[domain.Role]bool{
	domain.RoleOwner: true, domain.RoleEditor: true, domain.RoleViewer: true,
}

func ValidateRole(r domain.Role) error {
	if !roles[r] {
		return domain.Invalid("role", "must be one of owner, editor, viewer")
	}
	return nil
}

// CanRemoveMember guards the invariant that a project always keeps at least
// one owner, so it can never become unadministrable.
func CanRemoveMember(p domain.Project, user domain.UserID) error {
	role, ok := p.RoleOf(user)
	if !ok {
		return domain.Invalid("user", "is not a member of this project")
	}
	if role == domain.RoleOwner && len(p.Owners()) == 1 {
		return domain.Invalid("user", "cannot remove the last owner of a project")
	}
	return nil
}

var (
	nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)
	edgeDashes   = regexp.MustCompile(`^-+|-+$`)
)

// Slugify derives a kebab-case slug from free text, so a caller that supplies
// only a title or name still gets a stable, addressable identifier.
func Slugify(title string) string {
	s := nonSlugChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(title)), "-")
	s = edgeDashes.ReplaceAllString(s, "")
	if len(s) > SlugMaxLen {
		s = edgeDashes.ReplaceAllString(s[:SlugMaxLen], "")
	}
	return s
}
