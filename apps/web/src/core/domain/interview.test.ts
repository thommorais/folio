import { describe, expect, it } from 'vitest'
import {
	firstOpenQuestion,
	isExploring,
	isPending,
	isStuck,
	pendingActions,
	rounds,
	stageAnswer,
	stageDefer,
	stageExplore,
	stageReopen,
	stageThread,
	stagedCount,
	summarize,
	toActions,
	unstageDecision,
	unstageKind,
	unstageThreadAt,
	type Question,
	type Sent,
	type StagedMap,
} from './interview'

const question = (id: string, over: Partial<Question> = {}): Question => ({
	id,
	round: 1,
	deps: [],
	title: id,
	body: '',
	options: [
		{ k: 'a', text: 'A' },
		{ k: 'b', text: 'B' },
	],
	rec: { option: 'a', why: 'because' },
	status: 'open',
	durable: false,
	updated: false,
	thread: [],
	...over,
})

const empty: StagedMap = {}

describe('staging', () => {
	it('an answer replaces a staged defer or reopen on the same question', () => {
		const deferred = stageDefer(empty, 'q1')
		const answered = stageAnswer(deferred, 'q1', { kind: 'option', option: 'b' })

		expect(answered.q1).toEqual({ thread: [], answer: { kind: 'option', option: 'b' } })
	})

	it('a defer replaces a staged answer, and a reopen replaces a defer', () => {
		const answered = stageAnswer(empty, 'q1', { kind: 'text', text: 'x' })

		expect(stageDefer(answered, 'q1').q1?.answer).toBeUndefined()
		expect(stageReopen(stageDefer(empty, 'q1'), 'q1').q1?.defer).toBeUndefined()
	})

	it('keeps thread messages and explore when the decision changes', () => {
		let map = stageThread(empty, 'q1', 'why?')
		map = stageExplore(map, 'q1')
		map = stageAnswer(map, 'q1', { kind: 'accept', option: 'a' })
		map = stageDefer(map, 'q1')

		expect(map.q1).toEqual({ thread: ['why?'], explore: true, defer: true })
	})

	it('drops a question from the map once nothing is staged on it', () => {
		const map = unstageKind(stageExplore(empty, 'q1'), 'q1', 'explore')

		expect(map).toEqual({})
	})

	it('unstageDecision clears answer, defer and reopen but not the thread', () => {
		const map = unstageDecision(
			stageThread(stageAnswer(empty, 'q1', { kind: 'accept', option: 'a' }), 'q1', 'hi'),
			'q1',
		)

		expect(map.q1).toEqual({ thread: ['hi'] })
	})

	it('removes one thread message by index', () => {
		const map = unstageThreadAt(stageThread(stageThread(empty, 'q1', 'one'), 'q1', 'two'), 'q1', 0)

		expect(map.q1?.thread).toEqual(['two'])
	})
})

describe('toActions', () => {
	it('orders a question as reopen, defer, answer, explore, then its thread messages', () => {
		const map: StagedMap = {
			q1: {
				reopen: true,
				defer: true,
				answer: { kind: 'accept', option: 'a' },
				explore: true,
				thread: ['one', 'two'],
			},
		}

		expect(toActions(map, ['q1']).map(action => action.type)).toEqual([
			'reopen',
			'defer',
			'answer',
			'explore',
			'thread',
			'thread',
		])
	})

	it('follows the question order, not the order things were staged', () => {
		const map = stageAnswer(stageAnswer(empty, 'q2', { kind: 'option', option: 'b' }), 'q1', {
			kind: 'text',
			text: 'x',
		})

		expect(toActions(map, ['q1', 'q2']).map(action => ('q' in action ? action.q : ''))).toEqual(['q1', 'q2'])
	})

	it('carries the option on an accept and the text on a text answer', () => {
		const map = stageAnswer(stageAnswer(empty, 'q1', { kind: 'accept', option: 'a' }), 'q2', {
			kind: 'text',
			text: 'hm',
		})

		expect(toActions(map, ['q1', 'q2'])).toEqual([
			{ type: 'answer', q: 'q1', kind: 'accept', option: 'a' },
			{ type: 'answer', q: 'q2', kind: 'text', text: 'hm' },
		])
	})

	it('skips staged entries for questions that are no longer in the interview', () => {
		expect(toActions(stageDefer(empty, 'gone'), ['q1'])).toEqual([])
	})
})

describe('counting and summarising', () => {
	const map: StagedMap = {
		q1: { answer: { kind: 'accept', option: 'a' }, explore: true, thread: ['x', 'y'] },
		q2: { defer: true, thread: [] },
	}

	it('counts every staged action', () => {
		expect(stagedCount(map)).toBe(5)
	})

	it('summarises by kind with plurals', () => {
		expect(summarize(map)).toBe('1 answer, 2 messages, 1 deferral, 1 exploration')
	})
})

describe('rounds and selection', () => {
	const questions = [
		question('q1', { round: 1, status: 'answered' }),
		question('q2', { round: 1 }),
		question('q3', { round: 2, status: 'deferred' }),
		question('q4', { round: 2 }),
	]

	it('groups by round in ascending order', () => {
		expect(rounds(questions).map(([round, group]) => [round, group.map(q => q.id)])).toEqual([
			[1, ['q1', 'q2']],
			[2, ['q3', 'q4']],
		])
	})

	it('selects the first open question of the current round', () => {
		expect(firstOpenQuestion(questions)?.id).toBe('q4')
	})

	it('falls back to an open question of an earlier round, then to the first question', () => {
		expect(
			firstOpenQuestion([question('q1', { status: 'answered' }), question('q2', { status: 'reopened' })])?.id,
		).toBe('q2')
		expect(firstOpenQuestion([question('q1', { status: 'answered' })])?.id).toBe('q1')
		expect(firstOpenQuestion([])).toBeUndefined()
	})
})

describe('pending Sends', () => {
	const sent: Sent = {
		seq: 4,
		at: '2026-09-29T12:00:00Z',
		actions: [{ type: 'explore', q: 'q1' }, { type: 'thread', q: 'q2', text: 'hi' }, { type: 'finish' }],
	}

	it('is pending only while the agent has not handled its seq', () => {
		expect(isPending(sent, 3)).toBe(true)
		expect(isPending(sent, 4)).toBe(false)
		expect(isPending(undefined, 0)).toBe(false)
	})

	it('finds the actions on one question and ignores actions with no question', () => {
		expect(pendingActions(sent, 'q2')).toEqual([{ type: 'thread', q: 'q2', text: 'hi' }])
		expect(isExploring(sent, 'q1')).toBe(true)
		expect(isExploring(sent, 'q2')).toBe(false)
		expect(pendingActions(undefined, 'q1')).toEqual([])
	})

	it('treats an agent as stuck after five minutes', () => {
		const since = new Date('2026-09-29T12:00:00Z')

		expect(isStuck(since, since.getTime() + 4 * 60 * 1000)).toBe(false)
		expect(isStuck(since, since.getTime() + 5 * 60 * 1000 + 1)).toBe(true)
	})
})
