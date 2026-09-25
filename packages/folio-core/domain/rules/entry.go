package rules

import "folio/folio-core/domain"

var entryKinds = map[domain.EntryKind]bool{
	domain.EntryJournal: true, domain.EntryDoc: true, domain.EntryLog: true, domain.EntryResolution: true,
}

func ValidateEntry(e domain.Entry) error {
	if e.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if !entryKinds[e.Kind] {
		return domain.Invalid("kind", "must be one of journal, doc, log, resolution")
	}
	if e.Kind.Addressable() {
		if err := ValidateSlug("slug", e.Slug); err != nil {
			return err
		}
		if err := required("title", e.Title, TitleMaxLen); err != nil {
			return err
		}
	}
	if e.Kind == domain.EntryLog {
		if e.IssueID == "" && e.PlanID == "" {
			return domain.Invalid("issue", "a log must attach to an issue or a plan")
		}
		return required("body", e.Body, BodyMaxLen)
	}
	if e.Kind == domain.EntryResolution {
		if e.IssueID == "" {
			return domain.Invalid("issue", "a resolution must attach to an issue")
		}
		if e.PlanID != "" {
			return domain.Invalid("plan", "a resolution answers an issue, not a plan")
		}
		return required("body", e.Body, BodyMaxLen)
	}
	return optional("body", e.Body, BodyMaxLen)
}
