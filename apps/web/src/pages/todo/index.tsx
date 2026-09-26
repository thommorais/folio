import { Link, useParams } from '@tanstack/react-router'
import { useIssue } from '_/app/use-issue'
import { RecordGone } from '_/components/record/record-gone'
import { Status } from '_/lib/async-status'
import { TodoBody } from '_/pages/issues/issue-details'
import { LoadError } from '_/components/load-error'

const TodoDetail = () => {
	const { client, domain, slug, todo } = useParams({ from: '/_authenticated/$client/$domain/$slug/todos/$todo' })
	const state = useIssue(slug, todo)

	if (state.status === Status.Idle || state.status === Status.Loading) {
		return null
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
		return <LoadError message={state.message} />
	}

	return <TodoBody project={slug} todo={state.issue} />
}

export { TodoDetail }
