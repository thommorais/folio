import { describe, expect, it, vi } from 'vitest'
import { defaultChip } from './default-chip'

const ALL = ['open', 'done', 'cancelled'] as const
const DEFAULT = ['open'] as const
const label = (status: string) => status.toUpperCase()

describe('defaultChip', () => {
	it('names what the default hides while no status is chosen', () => {
		const chip = defaultChip('statuses', undefined, ALL, DEFAULT, label, vi.fn())

		expect(chip?.label).toBe('Hiding DONE, CANCELLED')
	})

	it('is absent once the reader picks statuses', () => {
		expect(defaultChip('statuses', ['done'], ALL, DEFAULT, label, vi.fn())).toBeUndefined()
	})

	it('shows everything when removed', () => {
		const showAll = vi.fn()

		defaultChip('statuses', undefined, ALL, DEFAULT, label, showAll)?.onRemove()

		expect(showAll).toHaveBeenCalledWith(ALL)
	})

	it('is absent when the default hides nothing', () => {
		expect(defaultChip('statuses', undefined, ALL, ALL, label, vi.fn())).toBeUndefined()
	})
})
