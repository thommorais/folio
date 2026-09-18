import { Badge } from '@thom/ui/badge'
import type { Member, Role } from '_/core/domain/domain'

const roleColor: Record<Role, 'active' | 'neutral' | 'muted'> = {
	owner: 'active',
	editor: 'neutral',
	viewer: 'muted',
}

const rank: Record<Role, number> = { owner: 0, editor: 1, viewer: 2 }

const label = (member: Member): string => member.name.trim() || member.email || 'Unknown'

export const MemberRoster = ({ members }: { readonly members: readonly Member[] }) => {
	if (members.length === 0) return <p className='text-dim text-sm'>No members yet.</p>

	const ordered = [...members].sort((a, b) => rank[a.role] - rank[b.role] || label(a).localeCompare(label(b)))

	return (
		<div className='space-y-2'>
			<h2 className='text-dim text-xs tracking-widest uppercase'>Members</h2>

			<ul className='flex flex-wrap gap-2'>
				{ordered.map(member => (
					<li key={member.userId}>
						<Badge color={roleColor[member.role]} title={member.email}>
							{label(member)}
							<span className='opacity-60'>{member.role}</span>
						</Badge>
					</li>
				))}
			</ul>
		</div>
	)
}
