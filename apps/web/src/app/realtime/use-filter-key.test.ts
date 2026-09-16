import { renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { filterKey, useFilterKey } from './use-filter-key'

describe('filterKey', () => {
	it('is the same string for equal filters written in a different key order', () => {
		expect(filterKey({ status: ['open'], ticketId: 'abc' })).toBe(filterKey({ ticketId: 'abc', status: ['open'] }))
	})

	it('treats an absent filter and an empty one alike', () => {
		expect(filterKey(undefined)).toBe(filterKey({}))
	})

	it('drops keys explicitly set to undefined, since the adapter ignores them', () => {
		expect(filterKey({ ticketId: 'abc', planId: undefined })).toBe(filterKey({ ticketId: 'abc' }))
	})

	it('keeps array order, which the adapter passes through to the filter', () => {
		expect(filterKey({ tags: ['a', 'b'] })).not.toBe(filterKey({ tags: ['b', 'a'] }))
	})

	it('separates filters that differ only in value', () => {
		expect(filterKey({ ticketId: 'a' })).not.toBe(filterKey({ ticketId: 'b' }))
	})

	it('survives a nested object', () => {
		const sort = { field: 'created', direction: 'desc' } as const
		expect(filterKey({ sort, ticketId: 'a' })).toBe(filterKey({ ticketId: 'a', sort }))
	})

	it('round-trips back to an equal filter', () => {
		const filter = { ticketId: 'abc', tags: ['x'], limit: 10 }
		expect(JSON.parse(filterKey(filter))).toEqual(filter)
	})
})

describe('useFilterKey', () => {
	it('keeps the same key across renders given an equal filter object', () => {
		const { result, rerender } = renderHook(({ filter }) => useFilterKey(filter), {
			initialProps: { filter: { ticketId: 'abc' } as Record<string, unknown> },
		})

		const before = result.current
		rerender({ filter: { ticketId: 'abc' } })

		expect(result.current).toBe(before)
	})

	it('changes the key when the filter changes', () => {
		const { result, rerender } = renderHook(({ filter }) => useFilterKey(filter), {
			initialProps: { filter: { ticketId: 'a' } as Record<string, unknown> },
		})

		const before = result.current
		rerender({ filter: { ticketId: 'b' } })

		expect(result.current).not.toBe(before)
	})
})
