import { useContainer } from '_/app/container'

export const LoadError = ({ message }: { readonly message: string }) => {
	const { connection } = useContainer()

	return (
		<div className='text-destructive flex flex-wrap items-center gap-3 text-sm'>
			<span>{message}</span>
			<button
				type='button'
				onClick={() => connection.retry()}
				className='text-foreground border-border hover:bg-accent/40 active:bg-accent/60 h-7 border px-2 text-xs transition-colors'
			>
				Retry
			</button>
		</div>
	)
}
