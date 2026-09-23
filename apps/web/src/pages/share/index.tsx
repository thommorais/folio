import { useParams } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { Heading } from '@thom/ui/heading'
import { useShared } from '_/app/use-shared'
import { Markdown } from '_/components/markdown'
import { Skeleton } from '_/components/motion/skeleton'
import type { SharedCycle, SharedEntry, SharedIssue, SharedItem, SharedPlan } from '_/core/domain/share'
import { Status } from '_/lib/async-status'
import { ISSUE_STATUS_LABELS } from '_/pages/issues/status-labels'
import { PLAN_STATUS_LABELS } from '_/pages/plans/status-labels'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const Section = ({ title, children }: { readonly title: string; readonly children: React.ReactNode }) => (
	<section className='space-y-2'>
		<h2 className='text-dim text-xs tracking-wide uppercase'>{title}</h2>
		{children}
	</section>
)

const Tags = ({ tags }: { readonly tags: readonly string[] }) =>
	tags.map(tag => (
		<Badge key={tag} color='muted'>
			{tag}
		</Badge>
	))

const TodoList = ({ todos }: { readonly todos: readonly (SharedIssue & { readonly id: string })[] }) => (
	<ul className='border-border divide-border divide-y border'>
		{todos.map(todo => (
			<li key={todo.id} className='flex items-center gap-3 px-4 py-3'>
				<span
					className={cn('border-border size-4 shrink-0 border', todo.status === 'done' && 'bg-foreground border-foreground')}
				/>
				<span className={cn('flex-1 truncate text-sm', todo.status === 'done' && 'text-dim line-through')}>
					{todo.title}
				</span>
				{todo.status === 'blocked' && <Badge color='destructive'>Blocked</Badge>}
				<span className='text-dim w-24 shrink-0 text-right text-xs'>{ISSUE_STATUS_LABELS[todo.status]}</span>
			</li>
		))}
	</ul>
)

const EntryList = ({ entries }: { readonly entries: readonly SharedEntry[] }) => (
	<ul className='border-border divide-border divide-y border'>
		{entries.map(entry => (
			<li key={entry.id} className='space-y-2 px-4 py-3'>
				<div className='flex items-baseline justify-between gap-4'>
					<span className='text-sm'>{entry.title}</span>
					<span className='text-dimmer shrink-0 text-xs'>{formatDate(entry.createdAt)}</span>
				</div>
				{entry.body && <Markdown>{entry.body}</Markdown>}
			</li>
		))}
	</ul>
)

const CycleList = ({ cycles }: { readonly cycles: readonly SharedCycle[] }) => (
	<ul className='border-border divide-border divide-y border'>
		{cycles.map(cycle => (
			<li key={cycle.id} className='flex items-baseline gap-3 px-4 py-3 text-sm'>
				<span className='text-dimmer font-mono text-xs'>#{cycle.ordinal}</span>
				<span className={cn('flex-1', !cycle.resolution && 'text-dim')}>{cycle.resolution || 'Open'}</span>
				<span className='text-dim shrink-0 text-xs'>{cycle.closedAt ? 'resolved' : cycle.phase}</span>
			</li>
		))}
	</ul>
)

const planProgress = (plan: SharedPlan): string =>
	plan.progress.total === 0 ? 'no todos' : `${plan.progress.done} of ${plan.progress.total} done`

const IssueView = ({ item }: { readonly item: Extract<SharedItem, { kind: 'issue' }> }) => {
	const { issue } = item

	return (
		<>
			<header className='space-y-3'>
				<Heading>{issue.title}</Heading>
				<div className='flex flex-wrap items-center gap-2'>
					<span className='text-dim text-xs'>{ISSUE_STATUS_LABELS[issue.status]}</span>
					<span className='text-dimmer font-mono text-xs'>{issue.priority}</span>
					{issue.externalRef && <span className='text-dimmer font-mono text-xs'>{issue.externalRef}</span>}
					<Tags tags={issue.tags} />
				</div>
				{issue.body && <Markdown>{issue.body}</Markdown>}
			</header>

			{item.cycles.length > 0 && (
				<Section title='Cycles'>
					<CycleList cycles={item.cycles} />
				</Section>
			)}

			{item.plans.length > 0 && (
				<Section title='Plans'>
					<ul className='border-border divide-border divide-y border'>
						{item.plans.map(plan => (
							<li key={plan.id} className='flex items-center justify-between gap-4 px-4 py-3 text-sm'>
								<span>{plan.title}</span>
								<span className='text-dim shrink-0 text-xs'>{PLAN_STATUS_LABELS[plan.status]}</span>
							</li>
						))}
					</ul>
				</Section>
			)}

			{item.todos.length > 0 && (
				<Section title='Todos'>
					<TodoList todos={item.todos} />
				</Section>
			)}

			{item.docs.length > 0 && (
				<Section title='Docs'>
					<EntryList entries={item.docs} />
				</Section>
			)}

			{item.journal.length > 0 && (
				<Section title='Journal'>
					<EntryList entries={item.journal} />
				</Section>
			)}
		</>
	)
}

const PlanView = ({ item }: { readonly item: Extract<SharedItem, { kind: 'plan' }> }) => {
	const { plan } = item

	return (
		<>
			<header className='space-y-3'>
				<Heading className={cn(plan.status === 'done' && 'text-dim line-through')}>{plan.title}</Heading>
				<div className='text-dimmer flex flex-wrap items-center gap-3 text-xs'>
					<Badge color='neutral'>{PLAN_STATUS_LABELS[plan.status]}</Badge>
					<span>{planProgress(plan)}</span>
					<span>Updated {formatDate(plan.updatedAt)}</span>
					<Tags tags={plan.tags} />
				</div>
			</header>

			{plan.goal ? <Markdown>{plan.goal}</Markdown> : <p className='text-dim text-sm'>This plan has no goal yet.</p>}

			{item.todos.length > 0 && (
				<Section title='Todos'>
					<TodoList todos={item.todos} />
				</Section>
			)}
		</>
	)
}

const Frame = ({ label, children }: { readonly label?: string; readonly children: React.ReactNode }) => (
	<main className='mx-auto flex min-h-dvh w-full max-w-3xl flex-col gap-8 px-4 py-10 sm:px-8'>
		<div className='text-dim flex items-center justify-between gap-4 text-xs'>
			<span className='font-serif text-sm'>folio</span>
			{label !== undefined && (
				<span className='flex items-center gap-2'>
					<span className='truncate'>Shared with {label}</span>
					<Badge color='muted'>Read-only</Badge>
				</span>
			)}
		</div>
		<article className='space-y-8'>{children}</article>
	</main>
)

const SharedPage = () => {
	const { token } = useParams({ from: '/share/$token' })
	const state = useShared(token)

	if (state.status === Status.Loading) {
		return (
			<Frame>
				<Skeleton className='h-32' />
			</Frame>
		)
	}

	if (state.status === Status.Gone) {
		return (
			<Frame>
				<div className='space-y-2 py-8'>
					<Heading>Link not found</Heading>
					<p className='text-dim text-sm'>This link is wrong or has been revoked. Ask whoever sent it for a new one.</p>
				</div>
			</Frame>
		)
	}

	if (state.status === Status.Failed) {
		return (
			<Frame>
				<p className='text-destructive text-sm'>{state.message}</p>
			</Frame>
		)
	}

	return (
		<Frame label={state.item.label}>
			{state.item.kind === 'issue' ? <IssueView item={state.item} /> : <PlanView item={state.item} />}
		</Frame>
	)
}

export { SharedPage }
