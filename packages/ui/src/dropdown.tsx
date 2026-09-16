'use client'

import { cn } from '@thom/libs/cn'
import * as Headless from '@headlessui/react'
import type React from 'react'
import { Button } from './button'
import { Link } from './link'

export function Dropdown(props: Headless.MenuProps) {
	return <Headless.Menu {...props} />
}

export function DropdownButton<T extends React.ElementType = typeof Button>({
	as = Button,
	...props
}: { className?: string } & Omit<Headless.MenuButtonProps<T>, 'className'>) {
	return <Headless.MenuButton as={as} {...props} />
}

export function DropdownMenu({
	anchor = 'bottom',
	className,
	...props
}: { className?: string } & Omit<Headless.MenuItemsProps, 'as' | 'className'>) {
	return (
		<Headless.MenuItems
			{...props}
			transition
			anchor={anchor}
			className={cn(
				'[--anchor-gap:--spacing(2)] [--anchor-padding:--spacing(1)]',
				'isolate w-max overflow-y-auto p-1',
				'border border-border bg-popover text-popover-foreground',
				'shadow-md',
				'focus:outline-hidden',
				'transition data-closed:data-leave:opacity-0 data-leave:duration-100 data-leave:ease-in',
				className,
			)}
		/>
	)
}

export function DropdownItem({
	className,
	...props
}: { className?: string } & (
	| ({ href?: never } & Omit<Headless.MenuItemProps<'button'>, 'as' | 'className'>)
	| ({ href: string } & Omit<Headless.MenuItemProps<typeof Link>, 'as' | 'className'>)
)) {
	const classes = cn(
		'group flex w-full cursor-default items-center gap-2 px-3 py-1.5',
		'text-left text-sm text-popover-foreground',
		'focus:outline-hidden data-focus:bg-accent',
		'data-disabled:opacity-50',
		'*:data-[slot=icon]:size-4 *:data-[slot=icon]:shrink-0 *:data-[slot=icon]:text-dim',
		className,
	)

	return typeof props.href === 'string' ? (
		<Headless.MenuItem as={Link} {...props} className={classes} />
	) : (
		<Headless.MenuItem as='button' type='button' {...props} className={classes} />
	)
}

export function DropdownHeader({ className, ...props }: React.ComponentPropsWithoutRef<'div'>) {
	return <div {...props} className={cn('px-3 pt-2 pb-1', className)} />
}

export function DropdownSection({
	className,
	...props
}: { className?: string } & Omit<Headless.MenuSectionProps, 'as' | 'className'>) {
	return <Headless.MenuSection {...props} className={className} />
}

export function DropdownHeading({
	className,
	...props
}: { className?: string } & Omit<Headless.MenuHeadingProps, 'as' | 'className'>) {
	return <Headless.MenuHeading {...props} className={cn('px-3 pt-2 pb-1 text-xs font-medium text-dim', className)} />
}

export function DropdownDivider({
	className,
	...props
}: { className?: string } & Omit<Headless.MenuSeparatorProps, 'as' | 'className'>) {
	return <Headless.MenuSeparator {...props} className={cn('my-1 h-px border-0 bg-border', className)} />
}

export function DropdownLabel({ className, ...props }: React.ComponentPropsWithoutRef<'div'>) {
	return <div {...props} data-slot='label' className={cn('truncate', className)} />
}

export function DropdownDescription({
	className,
	...props
}: { className?: string } & Omit<Headless.DescriptionProps, 'as' | 'className'>) {
	return <Headless.Description data-slot='description' {...props} className={cn('text-xs text-dimmer', className)} />
}

export function DropdownShortcut({
	keys,
	className,
	...props
}: { keys: string | string[]; className?: string } & Omit<Headless.DescriptionProps<'kbd'>, 'as' | 'className'>) {
	return (
		<Headless.Description as='kbd' {...props} className={cn('ml-auto flex text-xs text-dim', className)}>
			{(Array.isArray(keys) ? keys : keys.split('')).map(char => (
				<kbd key={char} className='min-w-[2ch] text-center font-sans capitalize'>
					{char}
				</kbd>
			))}
		</Headless.Description>
	)
}
