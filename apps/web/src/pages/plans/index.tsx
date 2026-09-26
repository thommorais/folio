import { Link, useParams, useSearch } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Tag } from '_/components/issue/tag'
import { StaggerItem } from '_/components/motion/stagger'
import { Flash } from '_/components/motion/flash'
import { MarkdownPreview } from '_/components/markdown/preview'
import { EmptyState } from '_/components/empty-state'
import { usePlans } from '_/app/use-plans'
import { DEFAULT_PLAN_STATUSES, PLAN_STATUS, type Plan } from '_/core/domain/plan'
import { PlanFilters } from './plan-filters'
import { PLAN_STATUS_LABELS } from './status-labels'
import { useScope } from '_/routing/use-scope'
import { Status } from '_/lib/async-status'
import { LoadError } from '_/components/load-error'

const Row = ({ plan, project }: { readonly plan: Plan; readonly project: string }) => {
	const { client, domain } = useScope()

	return (
		<Link
			to='/$client/$domain/$slug/plans/$plan'
			params={{ client, domain, slug: project, plan: plan.id }}
			data-row
			className='hover:bg-accent/40 active:bg-accent/60 relative block space-y-2 px-4 py-4 transition-colors'
		>
			<Flash on={plan.updatedAt.getTime()} />
			<div className='flex items-start justify-between gap-4'>
				<h3 className={cn('text-sm font-medium', plan.status === PLAN_STATUS.DONE && 'text-dim line-through')}>{plan.title}</h3>
				<span className='text-dim shrink-0 text-xs'>{PLAN_STATUS_LABELS[plan.status]}</span>
			</div>

			{plan.goal && <MarkdownPreview>{plan.goal}</MarkdownPreview>}

			{plan.tags.length > 0 && (
				<div className='flex flex-wrap items-center gap-2 pt-1'>
					{plan.tags.map(tag => (
						<Tag key={tag} tag={tag} />
					))}
				</div>
			)}
		</Link>
	)
}

const Plans = () => {
	const { slug } = useParams({ from: '/_authenticated/$client/$domain/$slug/plans/' })
	const search = useSearch({ from: '/_authenticated/$client/$domain/$slug/plans/' })

	const state = usePlans(slug, {
		ticketId: search.ticket,
		status: search.statuses ?? DEFAULT_PLAN_STATUSES,
		tags: search.tags,
		search: search.q,
		sort: search.sort,
	})

	const filtered =
		search.q !== undefined || search.statuses !== undefined || search.tags !== undefined || search.ticket !== undefined

	return (
		<div className='space-y-4'>
			<PlanFilters />

			{state.status === Status.Failed && <LoadError message={state.message} />}

			{state.status === Status.Ready && state.plans.length === 0 && filtered && (
				<p className='text-dim text-sm'>No plans match.</p>
			)}

			{state.status === Status.Ready && state.plans.length === 0 && !filtered && (
				<EmptyState message='No active plans.' command={`folio plan create "<title>" -p ${slug}`} />
			)}

			{state.status === Status.Ready && state.plans.length > 0 && (
				<ul className='border-border divide-border divide-y border'>
					{state.plans.map((plan, index) => (
						<StaggerItem key={plan.id} index={index} as='li'>
							<Row plan={plan} project={slug} />
						</StaggerItem>
					))}
				</ul>
			)}
		</div>
	)
}

export { Plans }
