import { useCallback, useEffect, useState } from 'react'
import { useContainer } from './container'
import {
	isPending,
	stagedCount,
	toActions,
	type Interview,
	type Sent,
	type SendAction,
	type StagedMap,
} from '_/core/domain/interview'

type Drafts = Readonly<Record<string, { readonly text?: string; readonly thread?: string }>>

type Stored = {
	readonly staged: StagedMap
	readonly sent?: Sent
	readonly drafts: Drafts
}

const blank: Stored = { staged: {}, drafts: {} }

const keyOf = (id: string): string => `folio:interview:${id}`

const read = (id: string): Stored => {
	try {
		const raw = window.localStorage.getItem(keyOf(id))

		return raw ? { ...blank, ...(JSON.parse(raw) as Partial<Stored>) } : blank
	} catch {
		return blank
	}
}

const write = (id: string, value: Stored): void => {
	try {
		window.localStorage.setItem(keyOf(id), JSON.stringify(value))
	} catch {
		return
	}
}

export const useInterviewSession = (interview: Interview) => {
	const { interviews } = useContainer()
	const { id, issueId, handled, questions } = interview
	const [stored, setStored] = useState<Stored>(() => read(id))
	const [sending, setSending] = useState(false)
	const [error, setError] = useState<string | undefined>(undefined)

	useEffect(() => {
		setStored(read(id))
	}, [id])

	useEffect(() => {
		write(id, stored)
	}, [id, stored])

	const setStaged = useCallback((change: (map: StagedMap) => StagedMap) => {
		setStored(current => ({ ...current, staged: change(current.staged) }))
	}, [])

	const setDraft = useCallback((q: string, field: 'text' | 'thread', value: string) => {
		setStored(current => ({
			...current,
			drafts: { ...current.drafts, [q]: { ...current.drafts[q], [field]: value } },
		}))
	}, [])

	const pending = isPending(stored.sent, handled) ? stored.sent : undefined
	const staged = stored.staged

	const send = useCallback(
		async (finish: boolean) => {
			const actions: SendAction[] = toActions(
				staged,
				questions.map(question => question.id),
			)

			if (finish) actions.push({ type: 'finish' })
			if (actions.length === 0 || sending) return

			setSending(true)
			setError(undefined)

			const result = await interviews.send(issueId, actions)

			setSending(false)

			if (!result.success) {
				setError(result.error.message)
				return
			}

			setStored(current => ({ ...current, staged: {}, sent: result.value }))
		},
		[interviews, issueId, questions, sending, staged],
	)

	return {
		staged,
		drafts: stored.drafts,
		pending,
		sending,
		error,
		count: stagedCount(staged),
		setStaged,
		setDraft,
		send,
	}
}
