import { cn } from '@thom/libs/cn'
import { Link } from './link'

export function Text({ className, ...props }: React.ComponentPropsWithoutRef<'p'>) {
	return <p data-slot='text' {...props} className={cn('text-sm text-dim', className)} data-id='thom-ui' />
}

export function TextLink({ className, ...props }: React.ComponentPropsWithoutRef<typeof Link>) {
	return <Link {...props} className={cn('text-foreground underline underline-offset-4', className)} data-id='thom-ui' />
}

export function Strong({ className, ...props }: React.ComponentPropsWithoutRef<'strong'>) {
	return <strong {...props} className={cn('font-medium text-foreground', className)} data-id='thom-ui' />
}

export function Code({ className, ...props }: React.ComponentPropsWithoutRef<'code'>) {
	return (
		<code
			{...props}
			className={cn('border border-border bg-accent px-1 py-0.5 text-xs font-medium text-foreground', className)}
			data-id='thom-ui'
		/>
	)
}
