import { Link, useParams } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { Heading } from '@thom/ui/heading'
import { Markdown } from '_/components/markdown'
import { Skeleton } from '_/components/motion/skeleton'
import { RecordGone } from '_/components/record/record-gone'
import { usePlan } from '_/app/use-plan'
import type { Plan, PlanStatus } from '_/core/domain/plan'
import { PLAN_STATUS_LABELS } from '_/pages/plans/status-labels'
import { useScope } from '_/routing/use-scope'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const statusColor = (status: PlanStatus) => {
	if (status === 'abandoned') return 'destructive' as const

	return status === 'done' ? ('active' as const) : ('neutral' as const)
}

const PlanBody = ({ plan, project }: { readonly plan: Plan; readonly project: string }) => {
	const { client, domain } = useScope()

	return (
		<article className='space-y-8'>
			<header className='space-y-3'>
				<Heading className={cn(plan.status === 'done' && 'text-dim line-through')}>{plan.title}</Heading>

				<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
					<Badge color={statusColor(plan.status)}>{PLAN_STATUS_LABELS[plan.status]}</Badge>
					<span>Updated {formatDate(plan.updatedAt)}</span>
					{plan.ticketId && (
						<Link
							to='/$client/$domain/$slug/tickets'
							params={{ client, domain, slug: project }}
							className='hover:text-foreground font-mono transition-colors'
						>
							{plan.ticketId}
						</Link>
					)}
					{plan.tags.map(tag => (
						<Badge key={tag} color='muted'>
							{tag}
						</Badge>
					))}
				</div>
			</header>

			{plan.goal ? <Markdown>{plan.goal}</Markdown> : <p className='text-dim text-sm'>This plan has no goal yet.</p>}
		</article>
	)
}

const PlanDetail = () => {
	const { client, domain, slug, plan } = useParams({ from: '/_authenticated/$client/$domain/$slug/plans/$plan' })
	const state = usePlan(slug, plan)

	if (state.status === 'idle' || state.status === 'loading') {
		return <Skeleton className='h-32' />
	}

	if (state.status === 'gone') {
		return (
			<RecordGone title={state.title}>
				<Link to='/$client/$domain/$slug/plans' params={{ client, domain, slug }} className='text-sm underline'>
					Back to plans
				</Link>
			</RecordGone>
		)
	}

	if (state.status === 'failed') {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	return <PlanBody plan={state.plan} project={slug} />
}

export { PlanDetail }
