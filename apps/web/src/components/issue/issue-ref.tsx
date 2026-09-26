import { Link } from '@tanstack/react-router'
import { useIssueById } from '_/app/use-issue'
import { ISSUE_KIND } from '_/core/domain/issue'
import { Status } from '_/lib/async-status'
import { useScope } from '_/routing/use-scope'
import { StatusIcon } from './status-icon'

export const IssueRef = ({ project, id }: { readonly project: string; readonly id: string }) => {
	const { client, domain } = useScope()
	const state = useIssueById(project, id)

	if (state.status !== Status.Ready) return null

	const { issue } = state

	return (
		<Link
			to={issue.kind === ISSUE_KIND.TODO ? '/$client/$domain/$slug/todos/$todo' : '/$client/$domain/$slug/tickets/$ticket'}
			params={{ client, domain, slug: project, ticket: issue.slug, todo: issue.slug }}
			className='hover:text-foreground flex min-w-0 items-center gap-1.5 transition-colors'
		>
			<StatusIcon status={issue.status} />
			<span className='truncate'>{issue.title}</span>
		</Link>
	)
}
