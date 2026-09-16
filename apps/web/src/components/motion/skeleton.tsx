import { cn } from '@thom/libs/cn'

export const Skeleton = ({ className }: { readonly className?: string }) => (
	<div className={cn('skeleton', className)} />
)
