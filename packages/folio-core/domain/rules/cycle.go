package rules

import (
	"strconv"
	"strings"

	"folio/folio-core/domain"
)

const ResolutionMaxLen = 2000

var phases = map[domain.Phase]bool{
	domain.PhasePlan: true, domain.PhaseDo: true,
	domain.PhaseCheck: true, domain.PhaseAct: true,
}

var phaseOrder = []domain.Phase{domain.PhasePlan, domain.PhaseDo, domain.PhaseCheck, domain.PhaseAct}

func ValidateCycle(c domain.Cycle) error {
	if c.ProjectID == "" {
		return domain.Invalid("project", "is required")
	}
	if c.IssueID == "" {
		return domain.Invalid("issue", "is required")
	}
	if c.Ordinal < 1 {
		return domain.Invalid("ordinal", "must be 1 or greater")
	}
	if !phases[c.Phase] {
		return domain.Invalid("phase", "must be one of plan, do, check, act")
	}
	return optional("resolution", c.Resolution, ResolutionMaxLen)
}

func CanTransitionPhase(from, to domain.Phase) error {
	if !phases[to] {
		return domain.Invalid("phase", "must be one of plan, do, check, act")
	}
	if from == to {
		return nil
	}
	for i, p := range phaseOrder {
		if p == from {
			if i+1 < len(phaseOrder) && phaseOrder[i+1] == to {
				return nil
			}
			break
		}
	}
	return domain.Invalid("phase", "must advance one step: plan, do, check, act")
}

func CurrentCycle(cycles []domain.Cycle) (domain.Cycle, bool) {
	var current domain.Cycle
	found := false
	for _, c := range cycles {
		if !found || c.Ordinal > current.Ordinal {
			current, found = c, true
		}
	}
	return current, found
}

func NextOrdinal(cycles []domain.Cycle) int {
	max := 0
	for _, c := range cycles {
		if c.Ordinal > max {
			max = c.Ordinal
		}
	}
	return max + 1
}

func CheckClosableIssue(status domain.IssueStatus, cycles []domain.Cycle) error {
	if !status.IsTerminal() {
		return nil
	}
	current, ok := CurrentCycle(cycles)
	if !ok {
		return nil
	}
	if strings.TrimSpace(current.Resolution) == "" {
		return domain.Invalid("resolution", "is required to close an issue: record what happened on cycle "+strconv.Itoa(current.Ordinal))
	}
	return nil
}

func CheckCycleMap(c domain.Cycle, m domain.Issue) error {
	if m.Kind != domain.IssueTicket || m.Wayfinder != domain.WayfinderMap {
		return domain.Invalid("map", "must be a ticket with wayfinder map")
	}
	if m.ProjectID != c.ProjectID {
		return domain.Invalid("map", "belongs to a different project")
	}
	if m.ID == c.IssueID {
		return domain.Invalid("map", "cannot be the ticket the cycle runs on")
	}
	if c.IsClosed() || c.Phase != domain.PhasePlan {
		return domain.Invalid("map", "can only be set while the cycle is in plan")
	}
	return nil
}

func CheckPlanClear(c domain.Cycle, to domain.Phase, children []domain.Issue) error {
	if c.MapID == "" || c.Phase != domain.PhasePlan || to == domain.PhasePlan {
		return nil
	}
	open := 0
	for _, child := range children {
		if child.Kind == domain.IssueTicket && !child.Status.IsTerminal() {
			open++
		}
	}
	if open > 0 {
		return domain.Invalid("phase", "plan holds while "+strconv.Itoa(open)+" decision(s) on the map are open")
	}
	return nil
}

func CheckCycleOwner(i domain.Issue) error {
	if i.Wayfinder == "" || i.Wayfinder == domain.WayfinderMap {
		return nil
	}
	return domain.Invalid("issue", "a "+string(i.Wayfinder)+" ticket is part of a plan; open the cycle on the ticket the plan is for")
}

func NextPhase(from domain.Phase) (domain.Phase, error) {
	for i, p := range phaseOrder {
		if p == from && i+1 < len(phaseOrder) {
			return phaseOrder[i+1], nil
		}
	}
	return "", domain.Invalid("phase", "act is the last phase; resolve the cycle instead")
}
