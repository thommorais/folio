import type { Unsubscribe } from '_/core/ports/subscription'
import { err, ok, type Result } from '_/lib/result'

type RecordEvent<TRecord> = { readonly action: string; readonly record: TRecord }

type Subscribable<TRecord> = {
	readonly subscribe: (topic: string, handler: (event: RecordEvent<TRecord>) => void) => Promise<Unsubscribe>
}

const message = (error: unknown): string => (error instanceof Error ? error.message : 'Unknown error')

export const subscribeToRecord = async <TRecord, TEntity>(
	collection: Subscribable<TRecord>,
	id: string,
	toEntity: (record: TRecord) => TEntity,
	onChange: (entity: TEntity) => void,
	onGone: () => void,
	label: string,
): Promise<Result<Unsubscribe>> => {
	try {
		const unsubscribe = await collection.subscribe(id, event => {
			if (event.action === 'delete') {
				onGone()
				return
			}

			onChange(toEntity(event.record))
		})

		return ok(unsubscribe)
	} catch (error) {
		return err(new Error(`Failed to subscribe to ${label} ${id}: ${message(error)}`))
	}
}
