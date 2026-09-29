import type { Answer, Explore, Interview, Message, Option, Question, Term } from '_/core/domain/interview'
import { interviewId as toInterviewId } from '_/core/domain/interview'
import type { InterviewsPort } from '_/core/ports/interviews'
import { err, ok, type Result } from '_/lib/result'
import { tryCatch } from '_/lib/try-catch'
import { z } from '_/lib/zod'
import { getPocketBaseClient } from './client'
import { keyed } from './request-key'
import { subscribeToRecord as subscribe } from './subscribe-to-record'

const COLLECTION = 'journ_interviews'
const BASE_PATH = '/api/folio'

const date = z.string().transform(value => new Date(value))

const messageSchema = z.object({
	who: z.enum(['user', 'agent']),
	text: z.string(),
	at: date.optional(),
})

const questionSchema = z.object({
	id: z.string(),
	round: z.number(),
	deps: z.array(z.string()).nullish(),
	title: z.string(),
	body: z.string().nullish(),
	options: z.array(z.object({ k: z.string(), text: z.string() })).nullish(),
	rec: z.object({ option: z.string().optional(), text: z.string().optional(), why: z.string().default('') }),
	status: z.enum(['open', 'answered', 'deferred', 'reopened']),
	durable: z.boolean().nullish(),
	updated: z.boolean().nullish(),
	answer: z
		.object({
			kind: z.enum(['accept', 'option', 'text']),
			option: z.string().optional(),
			text: z.string().optional(),
		})
		.nullish(),
	explore: z
		.object({
			at: date.optional(),
			rows: z.array(
				z.object({ option: z.string(), pros: z.array(z.string()).nullish(), cons: z.array(z.string()).nullish() }),
			),
		})
		.nullish(),
	thread: z.array(messageSchema).nullish(),
})

const stateSchema = z.object({
	note: z.string().nullish(),
	terms: z.array(z.object({ term: z.string(), def: z.string(), avoid: z.array(z.string()).nullish() })).nullish(),
	questions: z.array(questionSchema).nullish(),
})

type QuestionWire = z.infer<typeof questionSchema>

type InterviewRecord = {
	readonly id: string
	readonly issue: string
	readonly topic: string
	readonly state: unknown
	readonly agent_status: string
	readonly agent_since: string
	readonly handled: number
	readonly finished_at: string
	readonly created: string
}

const sentSchema = z.object({ seq: z.number(), at: z.string(), actions: z.array(z.any()).default([]) })

const epoch = new Date(0)

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

const toMessage = (wire: z.infer<typeof messageSchema>, fallback: Date): Message => ({
	who: wire.who,
	text: wire.text,
	at: wire.at ?? fallback,
})

const toQuestion = (wire: QuestionWire, fallback: Date): Question => ({
	id: wire.id,
	round: wire.round,
	deps: wire.deps ?? [],
	title: wire.title,
	body: wire.body ?? '',
	options: (wire.options ?? []) satisfies readonly Option[],
	rec: wire.rec,
	status: wire.status,
	durable: wire.durable ?? false,
	updated: wire.updated ?? false,
	answer: wire.answer ? (wire.answer satisfies Answer) : undefined,
	explore: wire.explore
		? ({
				at: wire.explore.at ?? fallback,
				rows: wire.explore.rows.map(row => ({ option: row.option, pros: row.pros ?? [], cons: row.cons ?? [] })),
			} satisfies Explore)
		: undefined,
	thread: (wire.thread ?? []).map(entry => toMessage(entry, fallback)),
})

const toInterview = (record: InterviewRecord): Result<Interview> => {
	const parsed = stateSchema.safeParse(record.state ?? {})

	if (!parsed.success) return err(new Error('The interview state has an unexpected shape.'))

	const created = new Date(record.created)
	const state = parsed.data

	return ok({
		id: toInterviewId(record.id),
		issueId: record.issue,
		topic: record.topic,
		note: state.note ?? '',
		terms: (state.terms ?? []).map((term): Term => ({ term: term.term, def: term.def, avoid: term.avoid ?? [] })),
		questions: (state.questions ?? []).map(question => toQuestion(question, created)),
		agentStatus: record.agent_status === 'working' ? 'working' : 'waiting',
		agentSince: record.agent_since ? new Date(record.agent_since) : epoch,
		handled: record.handled ?? 0,
		finishedAt: record.finished_at ? new Date(record.finished_at) : undefined,
	})
}

export const createInterviewsAdapter = (): InterviewsPort => {
	const client = getPocketBaseClient()
	const collection = () => client.collection(COLLECTION)

	return {
		current: async (issueId): Promise<Result<Interview>> => {
			const { data, error } = await tryCatch(
				collection().getList<InterviewRecord>(
					1,
					1,
					keyed('interviews.current', {
						filter: client.filter('issue = {:issue}', { issue: issueId }),
						sort: '-created',
					}),
				),
			)

			if (error) return err(new Error(`Failed to load the interview: ${error.message}`, { cause: error }))

			const record = data.items.at(0)

			return record ? toInterview(record) : err(new Error('This ticket has no interview yet.'))
		},

		subscribeToRecord: async (id, onChange, onGone): Promise<Result<() => Promise<void>>> =>
			subscribe<InterviewRecord, Result<Interview>>(
				collection(),
				id,
				toInterview,
				result => {
					if (result.success) onChange(result.value)
				},
				onGone,
				'interview',
			),

		send: async (issueId, actions) => {
			const { data, error } = await tryCatch(
				client.send(`${BASE_PATH}/issues/${encodeURIComponent(issueId)}/interview/sends`, {
					method: 'POST',
					body: { actions },
				}),
			)

			if (error) return err(new Error(`Could not send: ${message(error)}`, { cause: error }))

			const parsed = sentSchema.safeParse(data)

			return parsed.success
				? ok({ seq: parsed.data.seq, at: parsed.data.at, actions })
				: err(new Error('The server sent an unexpected answer.'))
		},
	}
}
