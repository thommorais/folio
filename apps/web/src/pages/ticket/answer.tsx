import { useEntries } from '_/app/use-entries'
import { Markdown } from '_/components/markdown'
import { ENTRY_KIND } from '_/core/domain/entry'
import type { Issue } from '_/core/domain/issue'
import { Status } from '_/lib/async-status'

type Props = {
	readonly project: string
	readonly ticket: Issue
}

const Answer = ({ project, ticket }: Props) => {
	const details = useEntries(project, { kind: ENTRY_KIND.RESOLUTION, issueId: ticket.id })
	const detail =
		details.status === Status.Ready ? details.entries.find(entry => entry.id === ticket.resolutionEntry) : undefined

	return (
		<section className='border-border border px-4 py-3'>
			<h2 className='text-dim mb-2 text-xs'>Answer</h2>
			<p className='text-base font-medium'>{ticket.resolution}</p>
			{detail?.body && (
				<div className='border-border mt-3 border-t pt-3 text-sm'>
					<Markdown>{detail.body}</Markdown>
				</div>
			)}
		</section>
	)
}

export { Answer }
