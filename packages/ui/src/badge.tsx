import type React from 'react'
import { tv, type VariantProps } from '@thom/libs/tv'

const badgeClasses = tv({
	base: 'inline-flex items-center gap-x-1.5 border px-1.5 py-0.5 text-xs font-medium',
	variants: {
		color: {
			neutral: 'border-border bg-accent text-foreground',
			muted: 'border-border text-dim bg-transparent',
			active: 'border-border bg-foreground text-background',
			destructive: 'border-destructive/20 bg-destructive/10 text-destructive',
		},
	},
	defaultVariants: { color: 'neutral' },
})

type BadgeProps = React.ComponentPropsWithoutRef<'span'> & VariantProps<typeof badgeClasses>

export function Badge({ color, className, ...props }: BadgeProps) {
	return <span {...props} className={badgeClasses({ color, class: className })} data-id='thom-ui' />
}
