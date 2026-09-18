export const WORK_TYPES = ['tickets', 'plans', 'todos'] as const

export type WorkType = (typeof WORK_TYPES)[number]

export const WORK_TYPE_LABELS: Record<WorkType, string> = {
	tickets: 'Tickets',
	plans: 'Plans',
	todos: 'Todos',
}
