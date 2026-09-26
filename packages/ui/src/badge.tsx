import type React from 'react'
import { tv, type VariantProps } from '@thom/libs/tv'

const badgeClasses = tv({
	base: 'inline-flex h-5 shrink-0 items-center gap-1.5 px-2 text-[11px] leading-none font-normal whitespace-nowrap',
	variants: {
		color: {
			neutral: 'bg-accent text-foreground',
			muted: 'bg-accent text-dim',
			active: 'bg-foreground text-background',
			destructive: 'bg-destructive/10 text-destructive',
		},
	},
	defaultVariants: { color: 'muted' },
})

type BadgeProps = React.ComponentPropsWithoutRef<'span'> & VariantProps<typeof badgeClasses>

export function Badge({ color, className, ...props }: BadgeProps) {
	return <span {...props} className={badgeClasses({ color, class: className })} data-id='thom-ui' />
}
