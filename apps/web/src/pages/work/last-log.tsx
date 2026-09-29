import { Link } from '@tanstack/react-router'
import { useEntries } from '_/app/use-entries'
import { useIssues } from '_/app/use-issues'
import { usePlans } from '_/app/use-plans'
import { LoadError } from '_/components/load-error'
import { MarkdownPreview } from '_/components/markdown/preview'
import { ENTRY_KIND } from '_/core/domain/entry'
import { logTarget, type LogTarget } from '_/core/domain/log-target'
import { Status } from '_/lib/async-status'

type Scope = {
	readonly client: string
	readonly domain: string
	readonly slug: string
}

const TargetLink = ({ scope, target }: { readonly scope: Scope; readonly target: LogTarget }) => {
	const className = 'hover:text-foreground transition-colors'

	if (target.kind === 'plan') {
		return (
			<Link to='/$client/$domain/$slug/plans/$plan' params={{ ...scope, plan: target.ref }} className={className}>
				{target.title}
			</Link>
		)
	}

	if (target.kind === 'todo') {
		return (
			<Link to='/$client/$domain/$slug/todos/$todo' params={{ ...scope, todo: target.ref }} className={className}>
				{target.title}
			</Link>
		)
	}

	return (
		<Link to='/$client/$domain/$slug/tickets/$ticket' params={{ ...scope, ticket: target.ref }} className={className}>
			{target.title}
		</Link>
	)
}

export const LastLog = ({ scope }: { readonly scope: Scope }) => {
	const logs = useEntries(scope.slug, { kind: ENTRY_KIND.LOG, limit: 1 })
	const issues = useIssues(scope.slug)
	const plans = usePlans(scope.slug)

	if (logs.status === Status.Failed) return <LoadError message={logs.message} />

	const entry = logs.status === Status.Ready ? logs.entries.at(0) : undefined
	if (entry === undefined) return null

	const target = logTarget(
		entry,
		issues.status === Status.Ready ? issues.issues : [],
		plans.status === Status.Ready ? plans.plans : [],
	)

	return (
		<section className='space-y-2'>
			<header className='text-dim text-xs tracking-widest uppercase'>Last log</header>
			<div className='border-border space-y-1 border px-4 py-3'>
				<MarkdownPreview className='line-clamp-3 text-sm'>{entry.body}</MarkdownPreview>
				<p className='text-dimmer flex gap-2 text-xs'>
					<span>{entry.createdAt.toLocaleString()}</span>
					{target !== undefined && (
						<>
							<span aria-hidden>·</span>
							<TargetLink scope={scope} target={target} />
						</>
					)}
				</p>
			</div>
		</section>
	)
}
