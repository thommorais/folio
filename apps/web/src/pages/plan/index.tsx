import { Link, useParams } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { Tag } from '_/components/issue/tag'
import { Heading } from '@thom/ui/heading'
import { Markdown } from '_/components/markdown'
import { RecordGone } from '_/components/record/record-gone'
import { ShareSheet } from '_/components/share/share-sheet'
import { usePlan } from '_/app/use-plan'
import { IssueRef } from '_/components/issue/issue-ref'
import { PLAN_STATUS, type Plan, type PlanStatus } from '_/core/domain/plan'
import { PLAN_STATUS_LABELS } from '_/pages/plans/status-labels'
import { Status } from '_/lib/async-status'
import { SHARE_KIND } from '_/core/domain/share'
import { LoadError } from '_/components/load-error'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const statusColor = (status: PlanStatus) => {
	if (status === PLAN_STATUS.ABANDONED) return 'destructive' as const

	return status === PLAN_STATUS.DONE ? ('active' as const) : ('neutral' as const)
}

const PlanBody = ({ plan, project }: { readonly plan: Plan; readonly project: string }) => {
	return (
		<article className='space-y-8'>
			<header className='space-y-3'>
				<div className='flex items-start justify-between gap-4'>
					<Heading className={cn(plan.status === PLAN_STATUS.DONE && 'text-dim line-through')}>{plan.title}</Heading>
					<ShareSheet target={{ kind: SHARE_KIND.PLAN, id: plan.id, projectId: plan.projectId }} />
				</div>

				<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
					<Badge color={statusColor(plan.status)}>{PLAN_STATUS_LABELS[plan.status]}</Badge>
					<span>Updated {formatDate(plan.updatedAt)}</span>
					{plan.ticketId && <IssueRef project={project} id={plan.ticketId} />}
					{plan.tags.map(tag => (
						<Tag key={tag} tag={tag} />
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

	if (state.status === Status.Idle || state.status === Status.Loading) {
		return null
	}

	if (state.status === Status.Gone) {
		return (
			<RecordGone title={state.title}>
				<Link to='/$client/$domain/$slug/plans' params={{ client, domain, slug }} className='text-sm underline'>
					Back to plans
				</Link>
			</RecordGone>
		)
	}

	if (state.status === Status.Failed) {
		return <LoadError message={state.message} />
	}

	return <PlanBody plan={state.plan} project={slug} />
}

export { PlanDetail }
