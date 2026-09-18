import { Badge } from '@thom/ui/badge'
import { RecordGone } from '_/components/record/record-gone'
import { cn } from '@thom/libs/cn'
import { useIssueById } from '_/app/use-issue'
import { useIssues } from '_/app/use-issues'
import { usePlans } from '_/app/use-plans'
import type { Issue, IssueStatus } from '_/core/domain/issue'
import { ISSUE_STATUS_LABELS } from './status-labels'

const Field = ({ label, children }: { readonly label: string; readonly children: React.ReactNode }) => (
	<div>
		<div className='text-dim mb-2 text-[12px]'>{label}</div>
		<div className='text-[14px]'>{children}</div>
	</div>
)

const Empty = () => <span className='text-dim'>-</span>

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const Skeleton = () => (
	<div className='space-y-6'>
		<div className='space-y-3'>
			<div className='bg-accent/40 h-4 w-24 animate-pulse' />
			<div className='bg-accent/40 h-6 w-3/5 animate-pulse' />
		</div>
		<div className='grid grid-cols-2 gap-4'>
			{['status', 'priority', 'ticket', 'plan'].map(key => (
				<div key={key} className='space-y-2'>
					<div className='bg-accent/40 h-3 w-16 animate-pulse' />
					<div className='bg-accent/40 h-4 w-24 animate-pulse' />
				</div>
			))}
		</div>
	</div>
)

const Body = ({ todo, project }: { readonly todo: Issue; readonly project: string }) => {
	// The record stores ids; a bare id tells the reader nothing, so each is
	// resolved to the title and falls back to the id if the lookup is not in yet.
	const tickets = useIssues(project, { kind: 'ticket' })
	const plans = usePlans(project)

	const ticketTitle =
		todo.parentId !== undefined && tickets.status === 'ready'
			? (tickets.issues.find(candidate => candidate.id === todo.parentId)?.title ?? todo.parentId)
			: todo.parentId

	const planTitle =
		todo.planId !== undefined && plans.status === 'ready'
			? (plans.plans.find(candidate => candidate.id === todo.planId)?.title ?? todo.planId)
			: todo.planId

	return (
		<div className='scrollbar-hide h-full overflow-auto pb-6'>
			<header className='mb-8'>
				<div className='text-dim flex items-center justify-between text-xs'>
					<span className='font-mono'>{todo.priority}</span>
					<span>{formatDate(todo.createdAt)}</span>
				</div>

				<h2 className={cn('mt-6 mb-3 text-lg', todo.status === 'done' && 'text-dim line-through')}>{todo.title}</h2>

				<div className='flex flex-wrap items-center gap-2'>
					<Badge color={statusColor(todo.status)}>{ISSUE_STATUS_LABELS[todo.status]}</Badge>
					{todo.tags.map(tag => (
						<Badge key={tag} color='muted'>
							{tag}
						</Badge>
					))}
				</div>
			</header>

			{todo.body && <div className='mb-6 border px-4 py-3 text-sm whitespace-pre-line'>{todo.body}</div>}

			<div className='grid grid-cols-2 gap-4'>
				<Field label='Ticket'>{ticketTitle ?? <Empty />}</Field>
				<Field label='Plan'>{planTitle ?? <Empty />}</Field>
				<Field label='Due'>{todo.dueDate ? formatDate(todo.dueDate) : <Empty />}</Field>
				<Field label='Position'>{todo.position}</Field>
				<Field label='Depends on'>
					{todo.dependsOn.length > 0 ? (
						`${todo.dependsOn.length} todo${todo.dependsOn.length === 1 ? '' : 's'}`
					) : (
						<Empty />
					)}
				</Field>
				<Field label='Updated'>{formatDate(todo.updatedAt)}</Field>
			</div>
		</div>
	)
}

const statusColor = (status: IssueStatus) => {
	if (status === 'blocked') {
		return 'destructive' as const
	}
	return status === 'done' ? ('active' as const) : ('neutral' as const)
}

type Props = {
	readonly project: string
	readonly todoId: string | undefined
}

const IssueDetails = ({ project, todoId }: Props) => {
	const state = useIssueById(project, todoId ?? '')

	if (state.status === 'gone') {
		return <RecordGone title={state.title} />
	}

	if (state.status === 'failed') {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	if (state.status !== 'ready') {
		return <Skeleton />
	}

	return <Body todo={state.issue} project={project} />
}

export { IssueDetails }
