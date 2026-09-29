import { useEntries } from '_/app/use-entries'
import { LoadError } from '_/components/load-error'
import { Markdown } from '_/components/markdown'
import { ENTRY_KIND } from '_/core/domain/entry'
import { Status } from '_/lib/async-status'

type Props = {
	readonly project: string
	readonly issueId?: string
	readonly planId?: string
}

export const WorkLog = ({ project, issueId, planId }: Props) => {
	const logs = useEntries(project, { kind: ENTRY_KIND.LOG, issueId, planId })

	return (
		<section className='space-y-2'>
			<h2 className='text-base font-medium'>Work log</h2>

			{logs.status === Status.Failed && <LoadError message={logs.message} />}

			{logs.status === Status.Ready && logs.entries.length === 0 && <p className='text-dim text-sm'>No logs yet.</p>}

			{logs.status === Status.Ready && logs.entries.length > 0 && (
				<ul className='border-border divide-border divide-y border'>
					{logs.entries.map(entry => (
						<li key={entry.id} className='space-y-1 px-4 py-3'>
							<Markdown>{entry.body}</Markdown>
							<p className='text-dimmer text-xs'>{entry.createdAt.toLocaleString()}</p>
						</li>
					))}
				</ul>
			)}
		</section>
	)
}
