package rules

import "folio/folio-core/domain"

func ValidateTicketLog(l domain.TicketLog) error {
	if l.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if l.IssueID == "" {
		return domain.Invalid("issue", "is required")
	}
	return required("body", l.Body, BodyMaxLen)
}

func ValidatePlanLog(l domain.PlanLog) error {
	if l.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if l.PlanID == "" {
		return domain.Invalid("plan", "is required")
	}
	return required("body", l.Body, BodyMaxLen)
}

func ValidateTodoLog(l domain.TodoLog) error {
	if l.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if l.IssueID == "" {
		return domain.Invalid("issue", "is required")
	}
	return required("body", l.Body, BodyMaxLen)
}
