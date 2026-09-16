import type { SearchHit } from '_/core/ports/search'
import { useEffect, useState } from 'react'
import { useContainer } from './container'

type SearchState = {
	readonly hits: readonly SearchHit[]
	readonly isFetching: boolean
}

export const useSearch = (term: string): SearchState => {
	const { search } = useContainer()
	const [state, setState] = useState<SearchState>({ hits: [], isFetching: false })

	useEffect(() => {
		if (term.trim() === '') {
			setState({ hits: [], isFetching: false })
			return
		}

		let cancelled = false
		setState(previous => ({ hits: previous.hits, isFetching: true }))

		void search.search({ text: term }).then(result => {
			if (cancelled) return
			setState({ hits: result.success ? result.value : [], isFetching: false })
		})

		return () => {
			cancelled = true
		}
	}, [term, search])

	return state
}
