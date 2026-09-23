import { cn } from '@thom/libs/cn'
import { useEntries } from '_/app/use-entries'
import type { Cycle } from '_/core/domain/cycle'
import { isResolved } from '_/core/domain/cycle'
import { cycleProgress, newestFirst } from '_/core/domain/cycle-progress'
import { Status } from '_/lib/async-status'

type Props = {
	readonly project: string
	readonly cycles: readonly Cycle[]
}

const PhaseRail = ({ cycle }: { readonly cycle: Cycle }) => (
	<ol className='flex items-center gap-1' aria-label={`Phase ${cycle.phase}`}>
		{cycleProgress(cycle).map((step, index) => (
			<li key={step.phase} className='flex items-center gap-1'>
				{index > 0 && <span className={cn('h-px w-3', step.reached ? 'bg-foreground' : 'bg-border')} aria-hidden />}
				<span
					className={cn(
						'border px-1.5 py-0.5 text-xs',
						step.reached ? 'border-foreground' : 'border-border text-dimmer',
						step.current && 'bg-foreground text-background',
						step.reached && !step.current && 'text-foreground',
					)}
					aria-current={step.current ? 'step' : undefined}
				>
					{step.phase}
				</span>
			</li>
		))}
	</ol>
)

// Work logs are stamped with the cycle that was open when they landed, so each
// round loads its own rather than the ticket's whole log being split up here.
const CycleBlock = ({ project, cycle }: { readonly project: string; readonly cycle: Cycle }) => {
	const logs = useEntries(project, { kind: 'log', cycleId: cycle.id })
	const resolved = isResolved(cycle)

	return (
		<li className='relative pl-6'>
			<span
				className={cn(
					'absolute top-1.5 left-0 size-2 border',
					resolved ? 'border-border bg-border' : 'border-foreground bg-background',
				)}
				aria-hidden
			/>

			<div className='space-y-3 pb-6'>
				<div className='flex flex-wrap items-center gap-x-3 gap-y-2'>
					<span className='text-dim font-mono text-xs'>cycle {cycle.ordinal}</span>
					<PhaseRail cycle={cycle} />
					{!resolved && <span className='text-dim text-xs'>open</span>}
				</div>

				{cycle.resolution && <p className='text-sm'>{cycle.resolution}</p>}

				{logs.status === Status.Ready && logs.entries.length > 0 && (
					<ul className='border-border space-y-3 border-l pl-4'>
						{logs.entries.map(entry => (
							<li key={entry.id} className='space-y-1'>
								<p className='text-sm whitespace-pre-line'>{entry.body}</p>
								<p className='text-dimmer text-xs'>{entry.createdAt.toLocaleString()}</p>
							</li>
						))}
					</ul>
				)}
			</div>
		</li>
	)
}

const CycleTimeline = ({ project, cycles }: Props) => {
	if (cycles.length === 0) return null

	return (
		// The rail sits behind the markers rather than between them, so a round
		// with a long work log does not break the line.
		<ol className='before:bg-border relative before:absolute before:top-2 before:bottom-4 before:left-[3px] before:w-px'>
			{newestFirst(cycles).map(cycle => (
				<CycleBlock key={cycle.id} project={project} cycle={cycle} />
			))}
		</ol>
	)
}

export { CycleTimeline }
