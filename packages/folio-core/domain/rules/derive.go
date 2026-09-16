package rules

import (
	"strings"

	"folio/folio-core/domain"
)

// ApplyBlocked recomputes Blocked for every todo against the full set, so
// dependency resolution is accurate. Blocked is transient: it is derived on
// read and never written back. A dependency that is absent from the set
// (deleted, or in another project) does not block.
func ApplyBlocked(todos []domain.Todo) {
	byID := make(map[domain.TodoID]domain.Todo, len(todos))
	for _, t := range todos {
		byID[t.ID] = t
	}
	for i, t := range todos {
		blocked := false
		for _, depID := range t.DependsOn {
			dep, ok := byID[depID]
			if ok && !dep.Status.IsTerminal() {
				blocked = true
				break
			}
		}
		todos[i].Blocked = blocked
	}
}

// ProgressOf counts a plan's completion. Cancelled todos leave the
// denominator: work that was called off should not hold a plan below 100%.
func ProgressOf(todos []domain.Todo) domain.Progress {
	var p domain.Progress
	for _, t := range todos {
		if t.Status == domain.TodoCancelled {
			continue
		}
		p.Total++
		if t.Status == domain.TodoDone {
			p.Done++
		}
	}
	return p
}

// CanTransitionTodo reports whether a status change is legal. The only
// forbidden move is reviving a cancelled todo: cancellation is a deliberate
// end state, and reopening one hides history that a new todo would keep.
func CanTransitionTodo(from, to domain.TodoStatus) error {
	if !todoStatuses[to] {
		return domain.Invalid("status", "must be one of pending, in_progress, done, blocked, cancelled")
	}
	if from == to {
		return nil
	}
	if from == domain.TodoCancelled {
		return domain.Invalid("status", "a cancelled todo cannot be reopened; create a new one instead")
	}
	return nil
}

// CanTransitionTicket reports whether a status change is legal. As with a
// todo, cancellation is the one end state that cannot be undone; closing is
// reversible because work reopens.
func CanTransitionTicket(from, to domain.TicketStatus) error {
	if !ticketStatuses[to] {
		return domain.Invalid("status", "must be one of open, in_progress, blocked, closed, cancelled")
	}
	if from == to {
		return nil
	}
	if from == domain.TicketCancelled {
		return domain.Invalid("status", "a cancelled ticket cannot be reopened; create a new one instead")
	}
	return nil
}

// NextPosition returns the position for a todo appended to the given set.
func NextPosition(todos []domain.Todo) int {
	max := 0
	for _, t := range todos {
		if t.Position > max {
			max = t.Position
		}
	}
	return max + 1
}

// Snippet trims text to max runes for a search result, cutting on a word
// boundary where one is close enough to the limit to be worth keeping.
func Snippet(text string, max int) string {
	clean := strings.Join(strings.Fields(text), " ")
	runes := []rune(clean)
	if len(runes) <= max {
		return clean
	}
	cut := string(runes[:max])
	if idx := strings.LastIndex(cut, " "); idx > max/2 {
		cut = cut[:idx]
	}
	return strings.TrimRight(cut, " ,.;:") + "…"
}

func ApplyTicketBlocked(tickets []domain.Ticket) {
	byID := make(map[domain.TicketID]domain.Ticket, len(tickets))
	for _, t := range tickets {
		byID[t.ID] = t
	}
	for i, t := range tickets {
		blocked := false
		for _, depID := range t.DependsOn {
			dep, ok := byID[depID]
			if ok && !dep.Status.IsTerminal() {
				blocked = true
				break
			}
		}
		tickets[i].Blocked = blocked
	}
}

func CheckNoTicketCycle(candidate domain.Ticket, set []domain.Ticket) error {
	deps := make(map[domain.TicketID][]domain.TicketID, len(set))
	for _, t := range set {
		deps[t.ID] = t.DependsOn
	}
	deps[candidate.ID] = candidate.DependsOn

	seen := make(map[domain.TicketID]bool, len(deps))
	var walk func(domain.TicketID) bool
	walk = func(id domain.TicketID) bool {
		if id == candidate.ID {
			return true
		}
		if seen[id] {
			return false
		}
		seen[id] = true
		for _, next := range deps[id] {
			if walk(next) {
				return true
			}
		}
		return false
	}

	for _, dep := range candidate.DependsOn {
		if walk(dep) {
			return domain.Invalid("depends_on", "introduces a dependency cycle")
		}
	}
	return nil
}

func CheckNoTicketAncestry(candidate domain.Ticket, set []domain.Ticket) error {
	parent := make(map[domain.TicketID]domain.TicketID, len(set))
	for _, t := range set {
		parent[t.ID] = t.ParentID
	}
	parent[candidate.ID] = candidate.ParentID

	seen := make(map[domain.TicketID]bool, len(parent))
	for id := candidate.ParentID; id != ""; id = parent[id] {
		if id == candidate.ID {
			return domain.Invalid("parent", "introduces a parent cycle")
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
	}
	return nil
}
