import { cn } from '@thom/libs/cn'
import { Badge } from '@thom/ui/badge'
import { tagTone } from './tag-tone'

export const Tag = ({ tag }: { readonly tag: string }) => (
	<Badge>
		<span aria-hidden className={cn('size-1.5', tagTone(tag))} />
		{tag}
	</Badge>
)

export const Tags = ({ tags, className }: { readonly tags: readonly string[]; readonly className?: string }) =>
	tags.length === 0 ? null : (
		<span className={cn('flex flex-wrap items-center gap-1', className)}>
			{tags.map(tag => (
				<Tag key={tag} tag={tag} />
			))}
		</span>
	)
