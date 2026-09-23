import { Link, useMatches, useParams } from '@tanstack/react-router'
import { House } from 'lucide-react'
import { useClient } from '_/app/use-client'
import { useDomain } from '_/app/use-domain'
import { useEntry } from '_/app/use-entry'

import { usePlan } from '_/app/use-plan'
import { useIssue } from '_/app/use-issue'
import { Status } from '_/lib/async-status'

const sectionLabels: Record<string, string> = {
	tickets: 'Tickets',
	plans: 'Plans',
	todos: 'Todos',
	journal: 'Journal',
}

type Crumb = {
	readonly key: string
	readonly label: React.ReactNode
	readonly title: string
	readonly to?: string
	readonly params?: Record<string, string>
}

const Separator = () => (
	<span aria-hidden className='select-none'>
		/
	</span>
)

// A record's title only exists after its fetch resolves, so a leaf crumb falls
// back to the slug already in the URL rather than collapsing the trail.
const useLeafLabel = (project: string | undefined, params: LeafParams) => {
	const ticket = useIssue(project ?? '', params.ticket ?? '')
	const entry = useEntry(project ?? '', params.entry ?? '')
	const plan = usePlan(project ?? '', params.plan)

	if (project === undefined) return undefined

	if (params.ticket !== undefined) {
		return ticket.status === Status.Ready ? ticket.issue.title : params.ticket
	}
	if (params.entry !== undefined) {
		return entry.status === Status.Ready ? entry.entry.title : params.entry
	}
	// A plan is addressed by id, which would read as noise in the trail, so it
	// waits for the title rather than falling back to the URL.
	if (params.plan !== undefined) {
		return plan.status === Status.Ready ? plan.plan.title : undefined
	}

	return undefined
}

type LeafParams = {
	readonly ticket: string | undefined
	readonly entry: string | undefined
	readonly plan: string | undefined
}

export const Breadcrumbs = () => {
	const matches = useMatches()
	const params = useParams({ strict: false })
	const client = typeof params.client === 'string' ? params.client : undefined
	const domain = typeof params.domain === 'string' ? params.domain : undefined
	const slug = typeof params.slug === 'string' ? params.slug : undefined
	const leaf: LeafParams = {
		ticket: typeof params.ticket === 'string' ? params.ticket : undefined,
		entry: typeof params.entry === 'string' ? params.entry : undefined,
		plan: typeof params.plan === 'string' ? params.plan : undefined,
	}
	const leafLabel = useLeafLabel(slug, leaf)
	const clientRecord = useClient(client)
	const domainRecord = useDomain(client, domain)

	const routeId = matches.at(-1)?.routeId ?? ''
	const section = Object.keys(sectionLabels).find(name => routeId.includes(`/$slug/${name}`))

	const crumbs: Crumb[] = [{ key: 'root', label: <House size={14} aria-label='Clients' />, title: 'Clients', to: '/' }]

	if (client !== undefined) {
		const label = clientRecord.status === Status.Ready ? clientRecord.client.name : client
		crumbs.push({ key: 'client', label, title: label, to: '/$client', params: { client } })
	}

	if (client !== undefined && domain !== undefined) {
		const label = domainRecord.status === Status.Ready ? domainRecord.domain.name : domain
		crumbs.push({
			key: 'domain',
			label,
			title: label,
			to: '/$client/$domain',
			params: { client, domain },
		})
	}

	if (client !== undefined && domain !== undefined && slug !== undefined) {
		crumbs.push({
			key: 'project',
			label: slug,
			title: slug,
			to: '/$client/$domain/$slug',
			params: { client, domain, slug },
		})
	}

	if (section !== undefined && client !== undefined && domain !== undefined && slug !== undefined) {
		crumbs.push({
			key: 'section',
			label: sectionLabels[section] as string,
			title: sectionLabels[section] as string,
			to: `/$client/$domain/$slug/${section}`,
			params: { client, domain, slug },
		})
	}

	if (leafLabel !== undefined) {
		crumbs.push({ key: 'leaf', label: leafLabel, title: leafLabel })
	}

	return (
		<nav aria-label='Breadcrumb' className='text-dim flex min-w-0 items-center gap-2 text-xs tracking-widest uppercase'>
			{crumbs.map((crumb, index) => {
				const isLast = index === crumbs.length - 1

				return (
					<span key={crumb.key} className='flex min-w-0 items-center gap-2'>
						{index > 0 && <Separator />}

						{isLast || crumb.to === undefined ? (
							<span className='text-foreground truncate' aria-current='page'>
								{crumb.label}
							</span>
						) : (
							<Link
								to={crumb.to}
								params={crumb.params}
								title={crumb.title}
								className='hover:text-foreground flex shrink-0 items-center transition-colors'
							>
								{crumb.label}
							</Link>
						)}
					</span>
				)
			})}
		</nav>
	)
}
