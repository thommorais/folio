import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'
import { toActions, type Question, type Sent, type StagedMap } from '_/core/domain/interview'
import { QuestionCard } from './question-card'

afterEach(cleanup)

const aQuestion = (over: Partial<Question> = {}): Question => ({
	id: 'q1',
	round: 1,
	deps: [],
	title: 'Tree or graph',
	body: '',
	options: [
		{ k: 'a', text: 'A tree' },
		{ k: 'b', text: 'A graph' },
	],
	rec: { option: 'b', why: 'Blockers show.' },
	status: 'open',
	durable: false,
	updated: false,
	thread: [],
	...over,
})

type HarnessProps = {
	readonly question: Question
	readonly sent?: Sent
	readonly locked?: boolean
	readonly onStaged?: (map: StagedMap) => void
}

const Harness = ({ question, sent, locked = false, onStaged }: HarnessProps) => {
	const [staged, setStaged] = useState<StagedMap>({})
	const [draft, setDraft] = useState('')

	onStaged?.(staged)

	return (
		<QuestionCard
			question={question}
			staged={staged[question.id]}
			sent={sent}
			draft={draft}
			locked={locked}
			onDraft={setDraft}
			onStage={change => setStaged(change)}
			onSelect={() => {}}
		/>
	)
}

const latest = () => {
	let map: StagedMap = {}

	return { record: (next: StagedMap) => (map = next), read: () => map }
}

describe('QuestionCard', () => {
	it('stages the recommended option as an accept that carries the option', async () => {
		const seen = latest()
		render(<Harness question={aQuestion()} onStaged={seen.record} />)

		await userEvent.click(screen.getByRole('button', { name: /A graph/ }))

		expect(toActions(seen.read(), ['q1'])).toEqual([{ type: 'answer', q: 'q1', kind: 'accept', option: 'b' }])
	})

	it('stages any other option as an option answer', async () => {
		const seen = latest()
		render(<Harness question={aQuestion()} onStaged={seen.record} />)

		await userEvent.click(screen.getByRole('button', { name: /A tree/ }))

		expect(toActions(seen.read(), ['q1'])).toEqual([{ type: 'answer', q: 'q1', kind: 'option', option: 'a' }])
	})

	it('unstages an option when it is clicked again', async () => {
		const seen = latest()
		render(<Harness question={aQuestion()} onStaged={seen.record} />)

		await userEvent.click(screen.getByRole('button', { name: /A tree/ }))
		await userEvent.click(screen.getByRole('button', { name: /A tree/ }))

		expect(seen.read()).toEqual({})
	})

	it('shows what is staged and clears it', async () => {
		const seen = latest()
		render(<Harness question={aQuestion()} onStaged={seen.record} />)

		await userEvent.click(screen.getByRole('button', { name: /A tree/ }))
		expect(screen.getByText(/Staged: option a/)).toBeDefined()

		await userEvent.click(screen.getByRole('button', { name: 'clear' }))

		expect(screen.queryByText(/Staged: option a/)).toBeNull()
		expect(seen.read()).toEqual({})
	})

	it('stages free text, empties the draft and offers to clear it', async () => {
		const seen = latest()
		render(<Harness question={aQuestion()} onStaged={seen.record} />)

		const box = screen.getByPlaceholderText('Free-text answer, if none of the options fit')
		await userEvent.type(box, 'Neither')
		await userEvent.click(screen.getByRole('button', { name: 'Stage answer' }))

		expect(toActions(seen.read(), ['q1'])).toEqual([{ type: 'answer', q: 'q1', kind: 'text', text: 'Neither' }])
		expect((box as HTMLTextAreaElement).value).toBe('')
		expect(screen.getByRole('button', { name: 'clear' })).toBeDefined()
	})

	it('stages the recommendation text of a question with no options as a text answer', async () => {
		const seen = latest()
		render(
			<Harness
				question={aQuestion({ options: [], rec: { text: 'waiting', why: 'Matches statuses.' } })}
				onStaged={seen.record}
			/>,
		)

		await userEvent.click(screen.getByRole('button', { name: 'Accept' }))

		expect(toActions(seen.read(), ['q1'])).toEqual([{ type: 'answer', q: 'q1', kind: 'text', text: 'waiting' }])
	})

	it('offers Defer on an open question and Reopen on a settled one', () => {
		const { unmount } = render(<Harness question={aQuestion()} />)
		expect(screen.getByRole('button', { name: 'Defer' })).toBeDefined()
		unmount()

		render(<Harness question={aQuestion({ status: 'answered', answer: { kind: 'accept', option: 'b' } })} />)
		expect(screen.getByRole('button', { name: 'Reopen' })).toBeDefined()
		expect(screen.queryByRole('button', { name: 'Defer' })).toBeNull()
	})

	it('lets an answered question be answered again', async () => {
		const seen = latest()
		render(
			<Harness
				question={aQuestion({ status: 'answered', answer: { kind: 'accept', option: 'b' } })}
				onStaged={seen.record}
			/>,
		)

		await userEvent.click(screen.getByRole('button', { name: /A tree/ }))

		expect(toActions(seen.read(), ['q1'])).toEqual([{ type: 'answer', q: 'q1', kind: 'option', option: 'a' }])
	})

	it('marks the option that was sent and is waiting on the agent', () => {
		const sent: Sent = {
			seq: 2,
			at: '2026-09-29T12:00:00Z',
			actions: [{ type: 'answer', q: 'q1', kind: 'option', option: 'a' }],
		}
		render(<Harness question={aQuestion()} sent={sent} />)

		expect(screen.getByRole('button', { name: /A tree.*sending/ })).toBeDefined()
	})

	it('says a sent defer is waiting and disables Explore while one is out', () => {
		const sent: Sent = {
			seq: 2,
			at: '2026-09-29T12:00:00Z',
			actions: [
				{ type: 'defer', q: 'q1' },
				{ type: 'explore', q: 'q1' },
			],
		}
		render(<Harness question={aQuestion()} sent={sent} />)

		expect(screen.getByText(/Sent: defer this question/)).toBeDefined()
		expect((screen.getByRole('button', { name: 'Exploring' }) as HTMLButtonElement).disabled).toBe(true)
	})

	it('disables staging and hides the answer box once the interview is finished', () => {
		render(<Harness question={aQuestion()} locked />)

		expect((screen.getByRole('button', { name: /A tree/ }) as HTMLButtonElement).disabled).toBe(true)
		expect(screen.queryByRole('button', { name: 'Stage answer' })).toBeNull()
	})

	it('labels Explore again once a table exists', () => {
		render(<Harness question={aQuestion({ explore: { at: new Date('2026-09-29T12:00:00Z'), rows: [] } })} />)

		expect(screen.getByRole('button', { name: 'Explore again' })).toBeDefined()
	})
})
