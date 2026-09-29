import type { Branded } from './branded'

export type InterviewId = Branded<string, 'InterviewId'>

export const interviewId = (value: string): InterviewId => value as InterviewId

export const QUESTION_STATUSES = ['open', 'answered', 'deferred', 'reopened'] as const
export type QuestionStatus = (typeof QUESTION_STATUSES)[number]

export type AgentStatus = 'waiting' | 'working'

export type AnswerKind = 'accept' | 'option' | 'text'

export type Option = { readonly k: string; readonly text: string }

export type Recommendation = {
	readonly option?: string
	readonly text?: string
	readonly why: string
}

export type Answer = {
	readonly kind: AnswerKind
	readonly option?: string
	readonly text?: string
}

export type ExploreRow = {
	readonly option: string
	readonly pros: readonly string[]
	readonly cons: readonly string[]
}

export type Explore = { readonly at: Date; readonly rows: readonly ExploreRow[] }

export type Message = { readonly who: 'user' | 'agent'; readonly text: string; readonly at: Date }

export type Question = {
	readonly id: string
	readonly round: number
	readonly deps: readonly string[]
	readonly title: string
	readonly body: string
	readonly options: readonly Option[]
	readonly rec: Recommendation
	readonly status: QuestionStatus
	readonly durable: boolean
	readonly updated: boolean
	readonly answer?: Answer
	readonly explore?: Explore
	readonly thread: readonly Message[]
}

export type Term = { readonly term: string; readonly def: string; readonly avoid: readonly string[] }

export type Interview = {
	readonly id: InterviewId
	readonly issueId: string
	readonly topic: string
	readonly note: string
	readonly terms: readonly Term[]
	readonly questions: readonly Question[]
	readonly agentStatus: AgentStatus
	readonly agentSince: Date
	readonly handled: number
	readonly finishedAt?: Date
}

export type SendAction =
	| {
			readonly type: 'answer'
			readonly q: string
			readonly kind: AnswerKind
			readonly option?: string
			readonly text?: string
	  }
	| { readonly type: 'thread'; readonly q: string; readonly text: string }
	| { readonly type: 'defer'; readonly q: string }
	| { readonly type: 'reopen'; readonly q: string }
	| { readonly type: 'explore'; readonly q: string }
	| { readonly type: 'finish' }

export type Sent = {
	readonly seq: number
	readonly at: string
	readonly actions: readonly SendAction[]
}

export type StagedAnswer = { readonly kind: AnswerKind; readonly option?: string; readonly text?: string }

export type Staged = {
	readonly answer?: StagedAnswer
	readonly defer?: true
	readonly reopen?: true
	readonly explore?: true
	readonly thread: readonly string[]
}

export type StagedMap = Readonly<Record<string, Staged>>

export const STUCK_AFTER_MS = 5 * 60 * 1000

export const isOpen = (question: Question): boolean => question.status === 'open' || question.status === 'reopened'

export const openCount = (questions: readonly Question[]): number => questions.filter(isOpen).length

export const currentRound = (questions: readonly Question[]): number =>
	questions.reduce((max, question) => Math.max(max, question.round), 0)

export const firstOpenQuestion = (questions: readonly Question[]): Question | undefined => {
	const round = currentRound(questions)

	return (
		questions.find(question => question.round === round && isOpen(question)) ??
		questions.find(isOpen) ??
		questions.at(0)
	)
}

export const rounds = (questions: readonly Question[]): ReadonlyArray<readonly [number, readonly Question[]]> => {
	const byRound = new Map<number, Question[]>()

	for (const question of questions) {
		byRound.set(question.round, [...(byRound.get(question.round) ?? []), question])
	}

	return [...byRound.entries()].sort(([a], [b]) => a - b)
}

const empty: Staged = { thread: [] }

const withEntry = (map: StagedMap, q: string, entry: Staged): StagedMap => {
	const { [q]: _dropped, ...rest } = map

	return isEmpty(entry) ? rest : { ...rest, [q]: entry }
}

export const isEmpty = (entry: Staged): boolean =>
	entry.answer === undefined && !entry.defer && !entry.reopen && !entry.explore && entry.thread.length === 0

export const stageAnswer = (map: StagedMap, q: string, answer: StagedAnswer): StagedMap => {
	const { defer: _defer, reopen: _reopen, ...rest } = map[q] ?? empty

	return withEntry(map, q, { ...rest, answer, thread: rest.thread })
}

export const stageDefer = (map: StagedMap, q: string): StagedMap => {
	const { answer: _answer, reopen: _reopen, ...rest } = map[q] ?? empty

	return withEntry(map, q, { ...rest, defer: true })
}

export const stageReopen = (map: StagedMap, q: string): StagedMap => {
	const { answer: _answer, defer: _defer, ...rest } = map[q] ?? empty

	return withEntry(map, q, { ...rest, reopen: true })
}

export const stageExplore = (map: StagedMap, q: string): StagedMap =>
	withEntry(map, q, { ...(map[q] ?? empty), explore: true })

export const stageThread = (map: StagedMap, q: string, text: string): StagedMap => {
	const entry = map[q] ?? empty

	return withEntry(map, q, { ...entry, thread: [...entry.thread, text] })
}

export const unstageKind = (map: StagedMap, q: string, kind: 'answer' | 'defer' | 'reopen' | 'explore'): StagedMap => {
	const { [kind]: _removed, ...rest } = map[q] ?? empty

	return withEntry(map, q, { ...rest, thread: rest.thread })
}

export const unstageThreadAt = (map: StagedMap, q: string, index: number): StagedMap => {
	const entry = map[q] ?? empty

	return withEntry(map, q, { ...entry, thread: entry.thread.filter((_, at) => at !== index) })
}

export const stagedCount = (map: StagedMap): number =>
	Object.values(map).reduce(
		(sum, entry) =>
			sum +
			(entry.answer ? 1 : 0) +
			(entry.defer ? 1 : 0) +
			(entry.reopen ? 1 : 0) +
			(entry.explore ? 1 : 0) +
			entry.thread.length,
		0,
	)

export const toActions = (map: StagedMap, order: readonly string[]): SendAction[] =>
	order.flatMap(q => {
		const entry = map[q]

		if (!entry) return []

		const actions: SendAction[] = []

		if (entry.reopen) actions.push({ type: 'reopen', q })
		if (entry.defer) actions.push({ type: 'defer', q })
		if (entry.answer) actions.push({ type: 'answer', q, ...entry.answer })
		if (entry.explore) actions.push({ type: 'explore', q })
		for (const text of entry.thread) actions.push({ type: 'thread', q, text })

		return actions
	})

export const summarize = (map: StagedMap): string => {
	const count = (pick: (entry: Staged) => number) => Object.values(map).reduce((sum, entry) => sum + pick(entry), 0)
	const parts = [
		[count(entry => (entry.answer ? 1 : 0)), 'answer'],
		[count(entry => entry.thread.length), 'message'],
		[count(entry => (entry.defer ? 1 : 0)), 'deferral'],
		[count(entry => (entry.reopen ? 1 : 0)), 'reopening'],
		[count(entry => (entry.explore ? 1 : 0)), 'exploration'],
	] as const

	return parts
		.filter(([n]) => n > 0)
		.map(([n, label]) => `${n} ${label}${n === 1 ? '' : 's'}`)
		.join(', ')
}

export const isPending = (sent: Sent | undefined, handled: number): sent is Sent =>
	sent !== undefined && sent.seq > handled

export const isStuck = (since: Date, now: number): boolean => now - since.getTime() > STUCK_AFTER_MS
