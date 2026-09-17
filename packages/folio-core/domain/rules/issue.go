package rules

import (
	"sort"

	"folio/folio-core/domain"
)

// ApplyLinks attaches each issue's relations and recomputes Blocked. A link to
// an issue outside the set does not block: an unresolvable dependency should
// not freeze work.
func ApplyLinks(issues []domain.Issue, links []domain.IssueLink) {
	byID := make(map[domain.IssueID]*domain.Issue, len(issues))
	for i := range issues {
		issues[i].ParentID = ""
		issues[i].DependsOn = nil
		issues[i].RelatedTo = nil
		issues[i].Blocked = false
		byID[issues[i].ID] = &issues[i]
	}

	for _, l := range links {
		switch l.Kind {
		case domain.LinkParent:
			if child, ok := byID[l.From]; ok {
				child.ParentID = l.To
			}
		case domain.LinkBlocks:
			if blocked, ok := byID[l.To]; ok {
				blocked.DependsOn = append(blocked.DependsOn, l.From)
			}
		case domain.LinkRelates:
			if a, ok := byID[l.From]; ok {
				a.RelatedTo = append(a.RelatedTo, l.To)
			}
			if b, ok := byID[l.To]; ok {
				b.RelatedTo = append(b.RelatedTo, l.From)
			}
		}
	}

	for _, issue := range byID {
		for _, dep := range issue.DependsOn {
			blocker, ok := byID[dep]
			if ok && !blocker.Status.IsTerminal() {
				issue.Blocked = true
				break
			}
		}
	}
}

// SortByScore orders issues by value per unit of effort, highest first. Ties
// keep a stable order by id so a listing does not shuffle between reads.
func SortByScore(issues []domain.Issue) {
	sort.SliceStable(issues, func(a, b int) bool {
		sa, sb := issues[a].Score(), issues[b].Score()
		if sa != sb {
			return sa > sb
		}
		return issues[a].ID < issues[b].ID
	})
}

// CheckNoLinkCycle rejects a link that would close a loop of the same kind.
func CheckNoLinkCycle(from, to domain.IssueID, kind domain.LinkKind, links []domain.IssueLink) error {
	if from == to {
		return domain.Invalid("link", "an issue cannot link to itself")
	}

	next := make(map[domain.IssueID][]domain.IssueID)
	for _, l := range links {
		if l.Kind == kind {
			next[l.From] = append(next[l.From], l.To)
		}
	}
	next[from] = append(next[from], to)

	seen := make(map[domain.IssueID]bool, len(next))
	var walk func(domain.IssueID) bool
	walk = func(id domain.IssueID) bool {
		if id == from {
			return true
		}
		if seen[id] {
			return false
		}
		seen[id] = true
		for _, n := range next[id] {
			if walk(n) {
				return true
			}
		}
		return false
	}

	if walk(to) {
		return domain.Invalid("link", "introduces a "+string(kind)+" cycle")
	}
	return nil
}

var issueStatuses = map[domain.IssueStatus]bool{
	domain.IssueOpen: true, domain.IssueInProgress: true, domain.IssueBlocked: true,
	domain.IssueDone: true, domain.IssueCancelled: true,
}

func CanTransitionIssue(from, to domain.IssueStatus) error {
	if !issueStatuses[to] {
		return domain.Invalid("status", "must be one of open, in_progress, blocked, done, cancelled")
	}
	if from == to {
		return nil
	}
	if from == domain.IssueCancelled {
		return domain.Invalid("status", "a cancelled issue cannot be reopened; create a new one instead")
	}
	return nil
}

func ValidateIssue(i domain.Issue) error {
	if i.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if i.Kind != domain.IssueTicket && i.Kind != domain.IssueTodo {
		return domain.Invalid("kind", "must be ticket or todo")
	}
	if err := ValidateSlug("slug", i.Slug); err != nil {
		return err
	}
	if err := required("title", i.Title, TitleMaxLen); err != nil {
		return err
	}
	if !issueStatuses[i.Status] {
		return domain.Invalid("status", "must be one of open, in_progress, blocked, done, cancelled")
	}
	if i.Size != 0 && !i.Size.Valid() {
		return domain.Invalid("size", "must be one of 1, 2, 3, 5, 8")
	}
	return nil
}

func ProgressOfIssues(issues []domain.Issue) domain.Progress {
	var p domain.Progress
	for _, i := range issues {
		if i.Status == domain.IssueCancelled {
			continue
		}
		p.Total++
		if i.Status == domain.IssueDone {
			p.Done++
		}
	}
	return p
}

func NextIssuePosition(issues []domain.Issue) int {
	max := 0
	for _, i := range issues {
		if i.Position > max {
			max = i.Position
		}
	}
	return max + 1
}
