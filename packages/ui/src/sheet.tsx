'use client'

import * as SheetPrimitive from '@radix-ui/react-dialog'
import type { ComponentPropsWithoutRef, ElementRef } from 'react'
import { forwardRef } from 'react'
import { tv } from '@thom/libs/tv'

const overlayClasses = tv({
	base: [
		'fixed',
		'inset-0',
		'z-50',
		'bg-background/60',
		'backdrop-blur-[2px]',
		'data-[state=open]:animate-[dialog-overlay-show_200ms_ease-out]',
		'data-[state=closed]:animate-[dialog-overlay-hide_150ms_ease-in]',
	],
})

const contentClasses = tv({
	base: ['fixed', 'inset-y-0', 'z-50', 'h-full', 'w-3/4', 'will-change-transform', 'md:p-4'],
	variants: {
		side: {
			left: [
				'left-0',
				'sm:max-w-sm',
				'data-[state=open]:animate-[sheet-in-left_320ms_cubic-bezier(0.16,1,0.3,1)]',
				'data-[state=closed]:animate-[sheet-out-left_200ms_cubic-bezier(0.4,0,1,1)]',
			],
			right: [
				'right-0',
				'sm:max-w-[520px]',
				'data-[state=open]:animate-[sheet-in-right_320ms_cubic-bezier(0.16,1,0.3,1)]',
				'data-[state=closed]:animate-[sheet-out-right_200ms_cubic-bezier(0.4,0,1,1)]',
			],
		},
	},
	defaultVariants: { side: 'right' },
})

const panelClasses = tv({
	base: ['bg-card', 'relative', 'h-full', 'w-full', 'overflow-hidden', 'border', 'p-6'],
})

const Sheet = SheetPrimitive.Root

const SheetTrigger = SheetPrimitive.Trigger

const SheetClose = SheetPrimitive.Close

type SheetContentProps = ComponentPropsWithoutRef<typeof SheetPrimitive.Content> & {
	readonly title: string
	readonly side?: 'left' | 'right'
}

const SheetContent = forwardRef<ElementRef<typeof SheetPrimitive.Content>, SheetContentProps>(
	({ className, children, title, side = 'right', ...props }, ref) => (
		<SheetPrimitive.Portal>
			<SheetPrimitive.Overlay className={overlayClasses()} />
			<SheetPrimitive.Content
				ref={ref}
				// Opening must not pull focus out of the list the sheet was opened from.
				onOpenAutoFocus={event => {
					event.preventDefault()
				}}
				className={contentClasses({ side })}
				{...props}
			>
				<div className={panelClasses({ class: className })}>
					<SheetPrimitive.Title className='sr-only'>{title}</SheetPrimitive.Title>
					{children}
				</div>
			</SheetPrimitive.Content>
		</SheetPrimitive.Portal>
	),
)
SheetContent.displayName = SheetPrimitive.Content.displayName

const headerClasses = tv({
	base: ['flex', 'flex-col', 'space-y-2'],
})

const footerClasses = tv({
	base: ['mt-auto', 'flex', 'shrink-0', 'justify-end', 'border-t', 'pt-4'],
})

const SheetHeader = ({ className, ...props }: React.ComponentPropsWithoutRef<'div'>) => (
	<div className={headerClasses({ class: className })} {...props} />
)

const SheetFooter = ({ className, ...props }: React.ComponentPropsWithoutRef<'div'>) => (
	<div className={footerClasses({ class: className })} {...props} />
)

export { Sheet, SheetClose, SheetContent, SheetFooter, SheetHeader, SheetTrigger }
