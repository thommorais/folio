'use client'
import type { ComponentPropsWithRef } from 'react'
import { tv, type VariantProps } from '@thom/libs/tv'

const buttonClasses = tv({
	base: [
		'inline-flex items-center justify-center gap-x-2',
		'text-sm font-medium',
		'transition-colors',
		'focus-visible:outline-none',
		'disabled:pointer-events-none disabled:opacity-50',
		'*:data-[slot=icon]:shrink-0',
	],
	variants: {
		variant: {
			solid: 'bg-primary text-primary-foreground hover:bg-primary/90',
			destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
			outline: 'border-border hover:bg-accent hover:text-accent-foreground border bg-transparent',
			secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
			ghost: 'hover:bg-accent hover:text-accent-foreground',
			link: 'text-primary underline-offset-4 hover:underline',
		},
		size: {
			sm: 'h-8 px-3 text-xs *:data-[slot=icon]:size-3.5',
			md: 'h-9 px-4 py-2 *:data-[slot=icon]:size-4',
			lg: 'h-10 px-8 *:data-[slot=icon]:size-4',
			icon: 'size-9 *:data-[slot=icon]:size-4',
		},
		fullWidth: { true: 'w-full' },
		loading: { true: 'cursor-wait' },
	},
	defaultVariants: {
		variant: 'solid',
		size: 'md',
	},
})

type ButtonProps = ComponentPropsWithRef<'button'> &
	VariantProps<typeof buttonClasses> & {
		loading?: boolean
	}

export const Button = ({ variant, size, fullWidth, loading, disabled, className, children, ...props }: ButtonProps) => {
	return (
		<button
			{...props}
			disabled={disabled || loading}
			className={buttonClasses({ variant, size, fullWidth, loading, class: className })}
			data-id='thom-ui'
		>
			<TouchTarget>
				{loading && <LoadingSpinner />}
				{children}
			</TouchTarget>
		</button>
	)
}

/** Expands the hit area to at least 44x44px on touch devices. */
export const TouchTarget = ({ children }: { children: React.ReactNode }) => {
	return (
		<>
			<span
				className='absolute top-1/2 left-1/2 size-[max(100%,2.75rem)] -translate-x-1/2 -translate-y-1/2 pointer-fine:hidden'
				aria-hidden='true'
			/>
			{children}
		</>
	)
}

const LoadingSpinner = () => {
	return (
		<svg
			className='size-4 animate-spin'
			xmlns='http://www.w3.org/2000/svg'
			fill='none'
			viewBox='0 0 24 24'
			data-slot='icon'
		>
			<circle className='opacity-25' cx='12' cy='12' r='10' stroke='currentColor' strokeWidth='4' />
			<path
				className='opacity-75'
				fill='currentColor'
				d='M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z'
			/>
		</svg>
	)
}
