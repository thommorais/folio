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
