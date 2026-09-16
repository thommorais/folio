import { Badge } from '@thom/ui/badge'
import { RecordGone } from '_/components/record/record-gone'
import { cn } from '@thom/libs/cn'
import { usePlan } from '_/app/use-plan'
import type { Plan, PlanStatus } from '_/core/domain/plan'
import { PLAN_STATUS_LABELS } from './status-labels'

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
			{['status', 'ticket', 'created', 'updated'].map(key => (
				<div key={key} className='space-y-2'>
					<div className='bg-accent/40 h-3 w-16 animate-pulse' />
					<div className='bg-accent/40 h-4 w-24 animate-pulse' />
				</div>
			))}
		</div>
	</div>
)

const statusColor = (status: PlanStatus) => {
	if (status === 'abandoned') {
		return 'destructive' as const
	}
	return status === 'done' ? ('active' as const) : ('neutral' as const)
}

const Body = ({ plan }: { readonly plan: Plan }) => (
	<div className='scrollbar-hide h-full overflow-auto pb-6'>
		<header className='mb-8'>
			<div className='text-dim flex items-center justify-between text-xs'>
				<span className='font-mono'>plan</span>
				<span>{formatDate(plan.createdAt)}</span>
			</div>

			<h2 className={cn('mt-6 mb-3 text-lg', plan.status === 'done' && 'text-dim line-through')}>{plan.title}</h2>

			<div className='flex flex-wrap items-center gap-2'>
				<Badge color={statusColor(plan.status)}>{PLAN_STATUS_LABELS[plan.status]}</Badge>
				{plan.tags.map(tag => (
					<Badge key={tag} color='muted'>
						{tag}
					</Badge>
				))}
			</div>
		</header>

		{plan.goal && <div className='mb-6 border px-4 py-3 text-sm whitespace-pre-line'>{plan.goal}</div>}

		<div className='grid grid-cols-2 gap-4'>
			<Field label='Ticket'>{plan.ticketId ?? <Empty />}</Field>
			<Field label='Status'>{PLAN_STATUS_LABELS[plan.status]}</Field>
			<Field label='Created'>{formatDate(plan.createdAt)}</Field>
			<Field label='Updated'>{formatDate(plan.updatedAt)}</Field>
		</div>
	</div>
)

type Props = {
	readonly project: string
	readonly planId: string | undefined
}

const PlanDetails = ({ project, planId }: Props) => {
	const state = usePlan(project, planId)

	if (state.status === 'gone') {
		return <RecordGone title={state.title} />
	}

	if (state.status === 'failed') {
		return <p className='text-destructive text-sm'>{state.message}</p>
	}

	if (state.status !== 'ready') {
		return <Skeleton />
	}

	return <Body plan={state.plan} />
}

export { PlanDetails }
