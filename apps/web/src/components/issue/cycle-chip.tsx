import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import type { Cycle } from '_/core/domain/cycle'
import { cycleProgress } from '_/core/domain/cycle-progress'

export const CycleChip = ({ cycle }: { readonly cycle: Cycle }) => (
	<Badge className='gap-2 pr-0.5' title={`Cycle ${cycle.ordinal}, ${cycle.phase}`}>
		Cycle {cycle.ordinal}
		<span className='flex items-center gap-px'>
			{cycleProgress(cycle).map(step => (
				<span
					key={step.phase}
					className={cn(
						'flex size-4 items-center justify-center font-mono text-[10px] uppercase',
						step.current && 'bg-foreground text-background',
						step.reached && !step.current && 'text-foreground',
						!step.reached && 'text-dimmer',
					)}
				>
					{step.phase[0]}
				</span>
			))}
		</span>
	</Badge>
)
