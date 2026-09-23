import { useEffect, useState } from 'react'
import type { Knowledge } from '_/core/ports/knowledge'
import { Status } from '_/lib/async-status'
import { useContainer } from './container'

type ListState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly notes: readonly Knowledge[] }
	| { readonly status: typeof Status.Failed; readonly message: string }

type NoteState =
	| { readonly status: typeof Status.Loading }
	| { readonly status: typeof Status.Ready; readonly note: Knowledge }
	| { readonly status: typeof Status.Failed; readonly message: string }

// Knowledge is not subscribed to the way project records are: it has no
// project to scope a subscription by, and a note is read far more often than
// it changes.
export const useKnowledgeList = (text: string): ListState => {
	const { knowledge } = useContainer()
	const [state, setState] = useState<ListState>({ status: Status.Loading })

	useEffect(() => {
		let cancelled = false
		setState({ status: Status.Loading })

		void knowledge.list({ text }).then(result => {
			if (cancelled) return
			setState(
				result.success
					? { status: Status.Ready, notes: result.value }
					: { status: Status.Failed, message: result.error.message },
			)
		})

		return () => {
			cancelled = true
		}
	}, [knowledge, text])

	return state
}

export const useKnowledge = (ref: string): NoteState => {
	const { knowledge } = useContainer()
	const [state, setState] = useState<NoteState>({ status: Status.Loading })

	useEffect(() => {
		let cancelled = false
		setState({ status: Status.Loading })

		void knowledge.get(ref).then(result => {
			if (cancelled) return
			setState(
				result.success
					? { status: Status.Ready, note: result.value }
					: { status: Status.Failed, message: result.error.message },
			)
		})

		return () => {
			cancelled = true
		}
	}, [knowledge, ref])

	return state
}
