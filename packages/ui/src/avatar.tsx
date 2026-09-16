import { cn } from '@thom/libs/cn'
import type React from 'react'

type AvatarProps = {
	src?: string | null
	square?: boolean
	initials?: string
	alt?: string
	className?: string
}

export function Avatar({
	src = null,
	square = false,
	initials,
	alt = '',
	className,
	...props
}: AvatarProps & React.ComponentPropsWithoutRef<'span'>) {
	return (
		<span
			data-slot='avatar'
			{...props}
			className={cn(
				'inline-grid shrink-0 items-center justify-center bg-accent align-middle *:col-start-1 *:row-start-1',
				square ? 'rounded-none' : 'rounded-full',
				className,
			)}
		>
			{initials && (
				<span
					className='text-foreground text-xs font-medium uppercase select-none'
					aria-hidden={alt ? undefined : true}
				>
					{initials}
				</span>
			)}
			{src && <img className={cn('size-full', square ? 'rounded-none' : 'rounded-full')} src={src} alt={alt} />}
		</span>
	)
}
