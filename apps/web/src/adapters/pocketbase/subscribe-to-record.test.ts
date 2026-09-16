import { describe, expect, it, vi } from 'vitest'
import { subscribeToRecord } from './subscribe-to-record'

type Row = { readonly id: string; readonly title: string }

const fakeCollection = () => {
	const close = vi.fn(async () => {})
	let emit: (event: { action: string; record: { id: string; title: string } }) => void = () => {}

	return {
		close,
		emit: (action: string, record: { id: string; title: string }) => emit({ action, record }),
		subscribe: vi.fn(async (topic: string, handler: typeof emit) => {
			emit = handler
			return close
		}),
	}
}

const toRow = (record: { id: string; title: string }): Row => ({ id: record.id, title: record.title })

describe('subscribeToRecord', () => {
	it('subscribes to the record id as its own topic', async () => {
		const collection = fakeCollection()

		await subscribeToRecord(collection, 'abc', toRow, vi.fn(), vi.fn(), 'row')

		expect(collection.subscribe).toHaveBeenCalledWith('abc', expect.any(Function))
	})

	it('maps an update through the record mapper', async () => {
		const collection = fakeCollection()
		const onChange = vi.fn()

		await subscribeToRecord(collection, 'abc', toRow, onChange, vi.fn(), 'row')
		collection.emit('update', { id: 'abc', title: 'renamed' })

		expect(onChange).toHaveBeenCalledWith({ id: 'abc', title: 'renamed' })
	})

	it('reports a delete as gone rather than a change', async () => {
		const collection = fakeCollection()
		const onChange = vi.fn()
		const onGone = vi.fn()

		await subscribeToRecord(collection, 'abc', toRow, onChange, onGone, 'row')
		collection.emit('delete', { id: 'abc', title: 'doomed' })

		expect(onGone).toHaveBeenCalledTimes(1)
		expect(onChange).not.toHaveBeenCalled()
	})

	it('hands back the unsubscribe', async () => {
		const collection = fakeCollection()

		const result = await subscribeToRecord(collection, 'abc', toRow, vi.fn(), vi.fn(), 'row')

		expect(result.success).toBe(true)
		if (result.success) expect(result.value).toBe(collection.close)
	})

	it('fails with a named error when the subscription cannot open', async () => {
		const collection = {
			subscribe: vi.fn(async () => {
				throw new Error('offline')
			}),
		}

		const result = await subscribeToRecord(collection, 'abc', toRow, vi.fn(), vi.fn(), 'todo')

		expect(result.success).toBe(false)
		if (!result.success) expect(result.error.message).toBe('Failed to subscribe to todo abc: offline')
	})
})
