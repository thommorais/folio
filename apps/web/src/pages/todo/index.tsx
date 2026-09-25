import { Link, useParams } from '@tanstack/react-router'
import { useIssue } from '_/app/use-issue'
import { RecordGone } from '_/components/record/record-gone'
import { Status } from '_/lib/async-status'
import { TodoBody } from '_/pages/issues/issue-details'

const TodoDetail = () => {
	const { client, domain, slug, todo } = useParams({ from: '/_authenticated/$client/$domain/$slug/todos/$todo' })
	const state = useIssue(slug, todo)

	if (state.status === Status.Idle || state.status === Status.Loading) {
		return <div className='bg-accent/40 h-32 animate-pulse' />
	}

	if (state.status === Status.Gone) {
		return (
			<RecordGone title={state.title}>
				<Link to='/$client/$domain/$slug/todos' params={{ client, domain, slug }} className='text-sm underline'>
					Back to todos
				</Link>
			</RecordGone>
		)
	}

	if (state.status === Status.Failed) {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	return <TodoBody project={slug} todo={state.issue} />
}

export { TodoDetail }
