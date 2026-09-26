import { Link } from '@tanstack/react-router'
import { Tag } from '_/components/issue/tag'
import { PriorityIcon } from '_/components/issue/priority-icon'
import { StatusIcon } from '_/components/issue/status-icon'
import { IssueRef } from '_/components/issue/issue-ref'
import { PlanRef } from '_/components/issue/plan-ref'
import { RecordGone } from '_/components/record/record-gone'
import { Markdown } from '_/components/markdown'
import { cn } from '@thom/libs/cn'
import { useIssueById } from '_/app/use-issue'
import { ISSUE_STATUS, type Issue } from '_/core/domain/issue'
import { ISSUE_STATUS_LABELS } from './status-labels'
import { Status } from '_/lib/async-status'
import { useScope } from '_/routing/use-scope'
import { usePreviewStore } from '_/app/preview-store'

const Field = ({ label, children }: { readonly label: string; readonly children: React.ReactNode }) => (
	<div>
		<div className='text-dim mb-2 text-[12px]'>{label}</div>
		<div className='text-[14px]'>{children}</div>
	</div>
)

const Empty = () => <span className='text-dim'>-</span>

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const Body = ({ todo, project, linked = false }: { readonly todo: Issue; readonly project: string; readonly linked?: boolean }) => {
	const { client, domain } = useScope()
	const closePreview = usePreviewStore(state => state.closePreview)

	return (
		<div className='scrollbar-hide h-full overflow-auto pb-6'>
			<header className='mb-8'>
				<div className='text-dim flex items-center justify-between text-xs'>
					<span className='flex items-center gap-1.5 capitalize'>
						<PriorityIcon priority={todo.priority} />
						{todo.priority}
					</span>
					<span>{formatDate(todo.createdAt)}</span>
				</div>

				<h2 className={cn('mt-6 mb-3 text-lg', todo.status === ISSUE_STATUS.DONE && 'text-dim line-through')}>
					{linked ? (
						<Link
							to='/$client/$domain/$slug/todos/$todo'
							params={{ client, domain, slug: project, todo: todo.slug }}
							onClick={closePreview}
							className='hover:underline underline-offset-4'
						>
							{todo.title}
						</Link>
					) : (
						todo.title
					)}
				</h2>

				<div className='flex flex-wrap items-center gap-2'>
					<span className='text-dim flex items-center gap-1.5 text-xs'>
						<StatusIcon status={todo.status} />
						{ISSUE_STATUS_LABELS[todo.status]}
					</span>
					{todo.tags.map(tag => (
						<Tag key={tag} tag={tag} />
					))}
				</div>
			</header>

			{todo.body && (
				<div className='mb-6 border px-4 py-3'>
					<Markdown>{todo.body}</Markdown>
				</div>
			)}

			<div className='grid grid-cols-2 gap-4'>
				<Field label='Parent'>{todo.parentId ? <IssueRef project={project} id={todo.parentId} /> : <Empty />}</Field>
				<Field label='Plan'>{todo.planId ? <PlanRef project={project} id={todo.planId} /> : <Empty />}</Field>
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

type Props = {
	readonly project: string
	readonly todoId: string | undefined
}

const IssueDetails = ({ project, todoId }: Props) => {
	const state = useIssueById(project, todoId ?? '')

	if (state.status === Status.Gone) {
		return <RecordGone title={state.title} />
	}

	if (state.status === Status.Failed) {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	if (state.status !== Status.Ready) {
		return null
	}

	return <Body todo={state.issue} project={project} linked />
}

export { Body as TodoBody, IssueDetails }
