import { Link } from '@tanstack/react-router'
import type { GroupTarget } from '_/core/domain/grouping'
import { useScope } from '_/routing/use-scope'

type Props = {
	readonly target: GroupTarget | undefined
	readonly slug: string
}

const style = 'text-dim text-xs tracking-widest uppercase'

export const GroupHeader = ({ target, slug }: Props) => {
	const { client, domain } = useScope()

	if (target === undefined) return <span className={style}>No parent</span>

	if (target.kind === 'plan') {
		return (
			<span className={style}>
				<span className='text-dimmer mr-2'>Plan</span>
				<Link
					to='/$client/$domain/$slug/plans/$plan'
					params={{ client, domain, slug, plan: target.plan.id }}
					className='hover:text-foreground transition-colors'
				>
					{target.plan.title}
				</Link>
			</span>
		)
	}

	return (
		<span className={style}>
			<span className='text-dimmer mr-2'>Ticket</span>
			<Link
				to='/$client/$domain/$slug/tickets/$ticket'
				params={{ client, domain, slug, ticket: target.ticket.slug }}
				className='hover:text-foreground transition-colors'
			>
				{target.ticket.title}
			</Link>
		</span>
	)
}
