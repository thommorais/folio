import type { PlanStatus } from '_/core/domain/plan'

export const PLAN_STATUS_LABELS: Record<PlanStatus, string> = {
	draft: 'Draft',
	active: 'Active',
	done: 'Done',
	abandoned: 'Abandoned',
}
