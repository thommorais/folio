import { SHARE_KIND, type SharedIssue, type SharedItem, type SharedPlan } from '_/core/domain/share'
import type { SharesPort } from '_/core/ports/shares'
import { ENTRY_KINDS } from '_/core/domain/entry'
import { ISSUE_KINDS, ISSUE_STATUSES, PRIORITIES } from '_/core/domain/issue'
import { PHASES } from '_/core/domain/cycle'
import { PLAN_STATUSES } from '_/core/domain/plan'
import { ENVS } from '_/envs'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { z } from '_/lib/zod'

const date = z.string().transform(value => new Date(value))
const optionalDate = z
	.string()
	.optional()
	.transform(value => (value ? new Date(value) : undefined))

const issueSchema = z.object({
	id: z.string(),
	kind: z.enum(ISSUE_KINDS),
	title: z.string(),
	body: z.string().default(''),
	status: z.enum(ISSUE_STATUSES),
	priority: z.enum(PRIORITIES),
	tags: z.array(z.string()).default([]),
	external_ref: z.string().default(''),
	updated_at: date,
})

const planSchema = z.object({
	id: z.string(),
	title: z.string(),
	goal: z.string().default(''),
	status: z.enum(PLAN_STATUSES),
	tags: z.array(z.string()).default([]),
	progress: z.object({ total: z.number(), done: z.number() }),
	updated_at: date,
})

const entrySchema = z.object({
	id: z.string(),
	kind: z.enum(ENTRY_KINDS),
	title: z.string().default(''),
	body: z.string().default(''),
	created_at: date,
})

const cycleSchema = z.object({
	id: z.string(),
	ordinal: z.number(),
	phase: z.enum(PHASES),
	resolution: z.string().default(''),
	closed_at: optionalDate,
})

const responseSchema = z.discriminatedUnion('kind', [
	z.object({
		kind: z.literal(SHARE_KIND.ISSUE),
		label: z.string(),
		brief: z.object({
			issue: issueSchema,
			children: z.array(issueSchema).default([]),
			plans: z.array(planSchema).default([]),
			journal: z.array(entrySchema).default([]),
			docs: z.array(entrySchema).default([]),
			cycles: z.array(cycleSchema).default([]),
		}),
	}),
	z.object({
		kind: z.literal(SHARE_KIND.PLAN),
		label: z.string(),
		plan: planSchema,
		todos: z.array(issueSchema).default([]),
	}),
])

type IssueWire = z.infer<typeof issueSchema>
type PlanWire = z.infer<typeof planSchema>

const toIssue = (wire: IssueWire): SharedIssue & { readonly id: string } => ({
	id: wire.id,
	kind: wire.kind,
	title: wire.title,
	body: wire.body,
	status: wire.status,
	priority: wire.priority,
	tags: wire.tags,
	externalRef: wire.external_ref,
	updatedAt: wire.updated_at,
})

const toPlan = (wire: PlanWire): SharedPlan & { readonly id: string } => ({
	id: wire.id,
	title: wire.title,
	goal: wire.goal,
	status: wire.status,
	tags: wire.tags,
	progress: wire.progress,
	updatedAt: wire.updated_at,
})

const toItem = (wire: z.infer<typeof responseSchema>): SharedItem => {
	if (wire.kind === SHARE_KIND.PLAN) {
		return { kind: SHARE_KIND.PLAN, label: wire.label, plan: toPlan(wire.plan), todos: wire.todos.map(toIssue) }
	}

	const { brief } = wire

	return {
		kind: SHARE_KIND.ISSUE,
		label: wire.label,
		issue: toIssue(brief.issue),
		todos: brief.children.map(toIssue),
		plans: brief.plans.map(toPlan),
		journal: brief.journal.map(entry => ({ ...entry, createdAt: entry.created_at })),
		docs: brief.docs.map(entry => ({ ...entry, createdAt: entry.created_at })),
		cycles: brief.cycles.map(cycle => ({ ...cycle, closedAt: cycle.closed_at })),
	}
}

const shareUrl = (token: string): string => {
	const base = ENVS.PUBLIC_API_URL === '/' ? window.location.origin : ENVS.PUBLIC_API_URL

	return new URL(`/api/share/${encodeURIComponent(token)}`, base).toString()
}

export const createSharesAdapter = (): SharesPort => ({
	open: async (token): Promise<Result<SharedItem | undefined>> => {
		const { data: response, error } = await tryCatch(fetch(shareUrl(token), { credentials: 'omit' }))

		if (error) return err(new Error('Cannot reach the server.', { cause: error }))
		if (response.status === 404) return ok(undefined)
		if (!response.ok) return err(new Error(`The link could not be opened (${response.status}).`))

		const { data: body, error: bodyError } = await tryCatch(response.json())

		if (bodyError) return err(new Error('The server sent an unreadable answer.', { cause: bodyError }))

		const parsed = responseSchema.safeParse(body)

		return parsed.success ? ok(toItem(parsed.data)) : err(new Error('The server sent an unexpected answer.'))
	},
})
