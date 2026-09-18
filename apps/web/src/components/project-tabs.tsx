import { Link, useRouterState } from '@tanstack/react-router'
import { cn } from '@thom/libs/cn'

const tabs = [
	{ to: '/$client/$domain/$slug', label: 'Overview', exact: true },
	{ to: '/$client/$domain/$slug/tickets', label: 'Tickets', exact: false },
	{ to: '/$client/$domain/$slug/plans', label: 'Plans', exact: false },
	{ to: '/$client/$domain/$slug/todos', label: 'Todos', exact: false },
	{ to: '/$client/$domain/$slug/journal', label: 'Journal', exact: false },
	{ to: '/$client/$domain/$slug/docs', label: 'Docs', exact: false },
] as const

type Props = {
	readonly client: string
	readonly domain: string
	readonly slug: string
}

export const ProjectTabs = ({ client, domain, slug }: Props) => {
	const { location } = useRouterState()

	return (
		<nav className='border-border scrollbar-hide flex gap-6 overflow-x-auto border-b'>
			{tabs.map(({ to, label, exact }) => {
				const href = to.replace('$client', client).replace('$domain', domain).replace('$slug', slug)
				const isActive = exact ? location.pathname === href : location.pathname.startsWith(href)

				return (
					<Link
						key={to}
						to={to}
						params={{ client, domain, slug }}
						className={cn(
							'-mb-px shrink-0 border-b-2 pb-2 text-sm transition-colors',
							isActive ? 'border-foreground text-foreground' : 'text-dim hover:text-foreground border-transparent',
						)}
					>
						{label}
					</Link>
				)
			})}
		</nav>
	)
}
