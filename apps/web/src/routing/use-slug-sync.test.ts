import { renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { useSlugSync } from './use-slug-sync'

const navigate = vi.fn()

describe('useSlugSync', () => {
	it('stays put while the slug matches', () => {
		navigate.mockClear()

		renderHook(() =>
			useSlugSync({
				to: '/$slug/docs/$doc',
				params: { slug: 'p' },
				param: 'doc',
				current: 'a',
				record: { id: '1', slug: 'a' },
			}),
		)

		expect(navigate).not.toHaveBeenCalled()
	})

	it('rewrites the url when the record is renamed under it', () => {
		navigate.mockClear()

		const { rerender } = renderHook(({ record }) => useSlugSync({ current: 'a', record, rename: navigate }), {
			initialProps: { record: { id: '1', slug: 'a' } },
		})

		rerender({ record: { id: '1', slug: 'b' } })

		expect(navigate).toHaveBeenCalledWith('b')
	})

	it('does not drag the url back when the reader moves to another record', () => {
		navigate.mockClear()

		const { rerender } = renderHook(({ record, current }) => useSlugSync({ current, record, rename: navigate }), {
			initialProps: { record: { id: '1', slug: 'a' }, current: 'a' },
		})

		// The url moves to doc b while the hook still holds doc a: a param change
		// reuses the component, so the old record outlives the old url by a render.
		rerender({ record: { id: '1', slug: 'a' }, current: 'b' })

		expect(navigate).not.toHaveBeenCalled()

		rerender({ record: { id: '2', slug: 'b' }, current: 'b' })

		expect(navigate).not.toHaveBeenCalled()
	})

	it('waits for a record before deciding anything', () => {
		navigate.mockClear()

		renderHook(() => useSlugSync({ current: 'a', record: undefined, rename: navigate }))

		expect(navigate).not.toHaveBeenCalled()
	})

	it('navigates once for one rename, not on every render', () => {
		navigate.mockClear()

		const { rerender } = renderHook(({ record }) => useSlugSync({ current: 'a', record, rename: navigate }), {
			initialProps: { record: { id: '1', slug: 'a' } },
		})

		rerender({ record: { id: '1', slug: 'b' } })
		rerender({ record: { id: '1', slug: 'b' } })

		expect(navigate).toHaveBeenCalledTimes(1)
	})
})
