import { useNavigate, useParams, useSearch } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { Sheet, SheetContent } from '@thom/ui/sheet'
import { Skeleton } from '_/components/motion/skeleton'
import { StaggerItem } from '_/components/motion/stagger'
import { usePlans } from '_/app/use-plans'
import type { Plan } from '_/core/domain/plan'
import type { PlansSearch } from '_/routes/_authenticated/$slug/plans'
import { PlanDetails } from './plan-details'
import { PlanFilters } from './plan-filters'
import { PLAN_STATUS_LABELS } from './status-labels'

const Row = ({ plan, onOpen }: { readonly plan: Plan; readonly onOpen: (id: string) => void }) => (
	<button
		type='button'
		onClick={() => {
			onOpen(plan.id)
		}}
		className='hover:bg-accent/40 w-full space-y-2 px-4 py-4 text-left transition-colors'
	>
		<div className='flex items-start justify-between gap-4'>
			<h3 className={cn('text-sm font-medium', plan.status === 'done' && 'text-dim line-through')}>{plan.title}</h3>
			<span className='text-dim shrink-0 text-xs'>{PLAN_STATUS_LABELS[plan.status]}</span>
		</div>

		{plan.goal && <p className='text-dim line-clamp-2 text-sm'>{plan.goal}</p>}

		{plan.tags.length > 0 && (
			<div className='flex flex-wrap items-center gap-2 pt-1'>
				{plan.tags.map(tag => (
					<Badge key={tag} color='muted'>
						{tag}
					</Badge>
				))}
			</div>
		)}
	</button>
)

const Plans = () => {
	const { slug } = useParams({ from: '/_authenticated/$slug/plans' })
	const search = useSearch({ from: '/_authenticated/$slug/plans' })
	const navigate = useNavigate()

	const state = usePlans(slug, {
		ticketId: search.ticket,
		status: search.statuses,
		tags: search.tags,
		search: search.q,
		sort: search.sort,
	})

	const setPlan = (plan: string | undefined) => {
		void navigate({ from: '/$slug/plans', to: '.', search: (prev: PlansSearch) => ({ ...prev, plan }) })
	}

	const filtered = search.q !== undefined || search.statuses !== undefined || search.tags !== undefined

	return (
		<div className='space-y-4'>
			<PlanFilters />

			{state.status === 'loading' && (
				<div className='border-border divide-border divide-y border'>
					{[0, 1].map(key => (
						<Skeleton key={key} className='h-20' />
					))}
				</div>
			)}

			{state.status === 'failed' && <p className='text-destructive text-sm'>{state.message}</p>}

			{state.status === 'ready' && state.plans.length === 0 && (
				<p className='text-dim text-sm'>{filtered ? 'No plans match.' : 'No plans yet.'}</p>
			)}

			{state.status === 'ready' && state.plans.length > 0 && (
				<ul className='border-border divide-border divide-y border'>
					{state.plans.map((plan, index) => (
						<StaggerItem key={plan.id} index={index} as='li'>
							<Row plan={plan} onOpen={setPlan} />
						</StaggerItem>
					))}
				</ul>
			)}

			<Sheet
				open={Boolean(search.plan)}
				onOpenChange={open => {
					if (!open) {
						setPlan(undefined)
					}
				}}
			>
				<SheetContent title='Plan details'>
					<PlanDetails project={slug} planId={search.plan} />
				</SheetContent>
			</Sheet>
		</div>
	)
}

export { Plans }
