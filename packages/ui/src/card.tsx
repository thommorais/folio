import type React from 'react'
import { tv, type VariantProps } from '@thom/libs/tv'

const cardClasses = tv({
	base: 'border-border bg-background text-card-foreground border',
	variants: {
		interactive: { true: 'hover:bg-accent/40 transition-colors' },
	},
})

const cardHeaderClasses = tv({ base: 'flex flex-col space-y-1.5 p-6' })
const cardContentClasses = tv({ base: 'p-6 pt-0' })
const cardFooterClasses = tv({ base: 'border-border text-dimmer flex items-center border-t p-6 text-xs' })
const cardTitleClasses = tv({ base: 'mb-2 text-lg leading-none font-medium tracking-tight' })
const cardDescriptionClasses = tv({ base: 'text-dimmer text-sm' })

type CardProps = React.ComponentPropsWithRef<'div'> & VariantProps<typeof cardClasses>

export const Card = ({ interactive, className, ...props }: CardProps) => (
	<div {...props} className={cardClasses({ interactive, class: className })} data-id='thom-ui' />
)

export const CardHeader = ({ className, ...props }: React.ComponentPropsWithRef<'div'>) => (
	<div {...props} className={cardHeaderClasses({ class: className })} />
)

export const CardContent = ({ className, ...props }: React.ComponentPropsWithRef<'div'>) => (
	<div {...props} className={cardContentClasses({ class: className })} />
)

export const CardFooter = ({ className, ...props }: React.ComponentPropsWithRef<'footer'>) => (
	<footer {...props} className={cardFooterClasses({ class: className })} />
)

export const CardTitle = ({ className, ...props }: React.ComponentPropsWithRef<'h3'>) => (
	<h3 {...props} className={cardTitleClasses({ class: className })} />
)

export const CardDescription = ({ className, ...props }: React.ComponentPropsWithRef<'p'>) => (
	<p {...props} className={cardDescriptionClasses({ class: className })} />
)
