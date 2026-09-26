import { useOnline } from '_/app/use-online'

export const ConnectionStatus = () => {
	const online = useOnline()

	if (online) return null

	return (
		<span role='status' className='text-dim flex items-center gap-2 text-xs'>
			<span aria-hidden className='size-1.5 animate-pulse bg-amber-500' />
			Reconnecting…
		</span>
	)
}
