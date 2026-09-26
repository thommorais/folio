import { cn } from '@thom/libs/cn'
import { useFlash } from './use-flash'

export const Flash = ({ on }: { readonly on: unknown }) => {
	const lit = useFlash(on)

	return (
		<span
			aria-hidden
			className={cn(
				'pointer-events-none absolute inset-0 bg-amber-500/10 transition-opacity',
				lit ? 'opacity-100 duration-150' : 'opacity-0 duration-1000',
			)}
		/>
	)
}
