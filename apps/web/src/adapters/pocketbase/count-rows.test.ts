import { describe, expect, it, vi } from 'vitest'
import { countRows } from './count-rows'

const collection = (totalItems: number) => ({
	getList: vi.fn().mockResolvedValue({ totalItems, items: [] }),
})

describe('countRows', () => {
	it('reads the total from the response rather than the returned rows', async () => {
		const rows = collection(137)

		await expect(countRows(rows, { filter: 'project.slug = {:p0}' })).resolves.toBe(137)
	})

	it('asks for a single row so the payload stays flat as a project grows', async () => {
		const rows = collection(137)

		await countRows(rows, { filter: 'project.slug = {:p0}' })

		expect(rows.getList).toHaveBeenCalledWith(1, 1, { filter: 'project.slug = {:p0}' })
	})
})
