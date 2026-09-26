import { Link } from '@tanstack/react-router'
import { usePlan } from '_/app/use-plan'
import { Status } from '_/lib/async-status'
import { useScope } from '_/routing/use-scope'

export const PlanRef = ({ project, id }: { readonly project: string; readonly id: string }) => {
	const { client, domain } = useScope()
	const state = usePlan(project, id)

	if (state.status !== Status.Ready) return null

	return (
		<Link
			to='/$client/$domain/$slug/plans/$plan'
			params={{ client, domain, slug: project, plan: state.plan.id }}
			className='hover:text-foreground block truncate transition-colors'
		>
			{state.plan.title}
		</Link>
	)
}
