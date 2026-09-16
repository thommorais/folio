'use client'

import * as DialogPrimitive from '@radix-ui/react-dialog'
import { Command as CommandPrimitive } from 'cmdk'
import type { ComponentPropsWithRef } from 'react'
import { cn } from '@thom/libs/cn'

export const Command = ({ className, ...props }: ComponentPropsWithRef<typeof CommandPrimitive>) => (
	<CommandPrimitive
		{...props}
		className={cn('text-popover-foreground flex h-full w-full flex-col overflow-hidden', className)}
	/>
)

export const CommandInput = ({ className, ...props }: ComponentPropsWithRef<typeof CommandPrimitive.Input>) => (
	<div className='flex w-full items-center' cmdk-input-wrapper=''>
		<CommandPrimitive.Input
			{...props}
			className={cn(
				'placeholder:text-muted-foreground flex h-10 w-full bg-transparent py-3 text-sm outline-none',
				'disabled:cursor-not-allowed disabled:opacity-50',
				className,
			)}
		/>
	</div>
)

export const CommandList = ({ className, ...props }: ComponentPropsWithRef<typeof CommandPrimitive.List>) => (
	<CommandPrimitive.List
		{...props}
		className={cn('max-h-[350px] w-full overflow-x-hidden overflow-y-auto', className)}
	/>
)

export const CommandEmpty = ({ className, ...props }: ComponentPropsWithRef<typeof CommandPrimitive.Empty>) => (
	<CommandPrimitive.Empty {...props} className={cn('py-6 text-center text-sm', className)} />
)

export const CommandGroup = ({ className, ...props }: ComponentPropsWithRef<typeof CommandPrimitive.Group>) => (
	<CommandPrimitive.Group
		{...props}
		className={cn(
			'text-foreground overflow-hidden p-1',
			'[&_[cmdk-group-heading]]:text-dim [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5',
			'[&_[cmdk-group-heading]]:text-[11px] [&_[cmdk-group-heading]]:font-medium',
			className,
		)}
	/>
)

export const CommandSeparator = ({ className, ...props }: ComponentPropsWithRef<typeof CommandPrimitive.Separator>) => (
	<CommandPrimitive.Separator {...props} className={cn('bg-border -mx-1 h-px', className)} />
)

export const CommandItem = ({ className, ...props }: ComponentPropsWithRef<typeof CommandPrimitive.Item>) => (
	<CommandPrimitive.Item
		{...props}
		className={cn(
			'relative flex cursor-default items-center px-2 py-1.5 text-sm outline-none select-none',
			'transition-colors duration-100',
			'aria-selected:bg-accent aria-selected:text-accent-foreground',
			className,
		)}
	/>
)

export const CommandShortcut = ({ className, ...props }: ComponentPropsWithRef<'span'>) => (
	<span {...props} className={cn('text-dim ml-auto text-xs tracking-widest', className)} />
)

export const CommandDialog = DialogPrimitive.Root

export const CommandDialogContent = ({
	className,
	children,
	...props
}: ComponentPropsWithRef<typeof DialogPrimitive.Content>) => (
	<DialogPrimitive.Portal>
		<DialogPrimitive.Overlay
			className={cn(
				// Literal colors: an opacity modifier on an hsl() token resolves to transparent in v4.
				'fixed inset-0 z-50 bg-[#f6f6f3]/60 backdrop-blur-[2px] dark:bg-[#0c0c0c]/80',
				'data-[state=closed]:animate-[dialog-overlay-hide_140ms_ease-in] data-[state=open]:animate-[dialog-overlay-show_180ms_ease-out]',
			)}
		/>
		<DialogPrimitive.Content
			{...props}
			className={cn(
				'text-foreground fixed top-1/2 left-1/2 z-50 -translate-x-1/2 -translate-y-1/2 will-change-transform select-text',
				'data-[state=closed]:animate-[dialog-content-hide_140ms_cubic-bezier(0.4,0,1,1)]',
				'data-[state=open]:animate-[dialog-content-show_240ms_cubic-bezier(0.16,1,0.3,1)]',
				className,
			)}
		>
			{children}
		</DialogPrimitive.Content>
	</DialogPrimitive.Portal>
)

export const CommandDialogTitle = DialogPrimitive.Title
