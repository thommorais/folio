import { describe, expect, it, vi } from 'vitest'

const getFullList = vi.fn()

vi.mock('./client', () => ({
	getPocketBaseClient: () => ({ collection: () => ({ getFullList }) }),
}))

const { linksOf, tagsByTarget } = await import('./relations')

describe('linksOf', () => {
	it('reads a parent link from the child that points at it', async () => {
		getFullList.mockResolvedValueOnce([{ from: 'child', to: 'parent', kind: 'parent' }])

		const links = await linksOf(['child'])

		expect(links.parentOf.get('child')).toBe('parent')
	})

	it('files a blocks link under the issue being blocked', async () => {
		getFullList.mockResolvedValueOnce([{ from: 'blocker', to: 'blocked', kind: 'blocks' }])

		const links = await linksOf(['blocked'])

		expect(links.dependsOn.get('blocked')).toEqual(['blocker'])
		expect(links.dependsOn.get('blocker')).toBeUndefined()
	})

	it('reads a relates link from both ends', async () => {
		getFullList.mockResolvedValueOnce([{ from: 'a', to: 'b', kind: 'relates' }])

		const links = await linksOf(['a', 'b'])

		expect(links.relatedTo.get('a')).toEqual(['b'])
		expect(links.relatedTo.get('b')).toEqual(['a'])
	})

	it('asks for nothing when there are no ids', async () => {
		getFullList.mockClear()

		await linksOf([])

		expect(getFullList).not.toHaveBeenCalled()
	})

	it('batches so a long id list cannot blow the filter length', async () => {
		getFullList.mockClear().mockResolvedValue([])

		await linksOf(Array.from({ length: 95 }, (_, i) => `id${i}`))

		expect(getFullList).toHaveBeenCalledTimes(3)
	})
})

describe('tagsByTarget', () => {
	it('groups tag names by the row they belong to', async () => {
		getFullList.mockResolvedValueOnce([
			{ issue: 'one', tag: 't1', expand: { tag: { name: 'backend' } } },
			{ issue: 'one', tag: 't2', expand: { tag: { name: 'search' } } },
			{ issue: 'two', tag: 't1', expand: { tag: { name: 'backend' } } },
		])

		const tags = await tagsByTarget('issue', ['one', 'two'])

		expect(tags.get('one')).toEqual(['backend', 'search'])
		expect(tags.get('two')).toEqual(['backend'])
	})

	it('skips a link whose tag could not be expanded', async () => {
		getFullList.mockResolvedValueOnce([{ issue: 'one', tag: 'gone', expand: {} }])

		const tags = await tagsByTarget('issue', ['one'])

		expect(tags.get('one')).toEqual([])
	})
})
