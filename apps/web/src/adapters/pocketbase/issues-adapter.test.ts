import { afterEach, describe, expect, it, vi } from 'vitest'

const send = vi.fn()

vi.mock('./client', () => ({
	getPocketBaseClient: () => ({ send }),
}))

const { createIssuesAdapter } = await import('./issues-adapter')

describe('issues adapter', () => {
	afterEach(() => {
		vi.resetAllMocks()
	})

	describe('update', () => {
		it('patches the editable fields', async () => {
			send.mockResolvedValueOnce({})

			const result = await createIssuesAdapter().update('i1', { title: 'T', body: 'B', priority: 'high', size: 5 })

			expect(send).toHaveBeenCalledWith('/api/folio/issues/i1', {
				method: 'PATCH',
				body: { title: 'T', body: 'B', priority: 'high', size: 5 },
			})
			expect(result).toEqual({ success: true, value: undefined })
		})

		it('sends size 0 to clear the size', async () => {
			send.mockResolvedValueOnce({})

			await createIssuesAdapter().update('i1', { title: 'T', body: '', priority: 'low', size: null })

			expect(send).toHaveBeenCalledWith(
				'/api/folio/issues/i1',
				expect.objectContaining({ body: expect.objectContaining({ size: 0 }) }),
			)
		})

		it('reports a failed save', async () => {
			send.mockRejectedValueOnce(new Error('title is required'))

			const result = await createIssuesAdapter().update('i1', { title: '', body: '', priority: 'low', size: null })

			expect(result).toMatchObject({
				success: false,
				error: { message: expect.stringContaining('Could not save the issue') },
			})
		})
	})

	describe('remove', () => {
		it('deletes through the service route so children are detached', async () => {
			send.mockResolvedValueOnce(undefined)

			const result = await createIssuesAdapter().remove('i1')

			expect(send).toHaveBeenCalledWith('/api/folio/issues/i1', { method: 'DELETE' })
			expect(result).toEqual({ success: true, value: undefined })
		})

		it('encodes the id in the path', async () => {
			send.mockResolvedValueOnce(undefined)

			await createIssuesAdapter().remove('a/b')

			expect(send).toHaveBeenCalledWith('/api/folio/issues/a%2Fb', { method: 'DELETE' })
		})

		it('reports a failed delete', async () => {
			send.mockRejectedValueOnce(new Error('forbidden'))

			const result = await createIssuesAdapter().remove('i1')

			expect(result).toMatchObject({
				success: false,
				error: { message: expect.stringContaining('Could not delete the issue') },
			})
		})
	})
})
