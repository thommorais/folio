import { Check, Copy } from 'lucide-react'
import { useCopy } from '_/app/use-copy'

const Command = ({ command }: { readonly command: string }) => {
	const { copied, copy } = useCopy()

	return (
		<button
			type='button'
			onClick={() => void copy(command)}
			aria-label={copied ? 'Copied' : `Copy ${command}`}
			className='bg-accent text-foreground hover:bg-accent/60 flex cursor-pointer items-center gap-2 px-1.5 py-0.5 font-mono transition-colors'
		>
			{command}
			{copied ? <Check size={12} className='text-dim' /> : <Copy size={12} className='text-dim' />}
		</button>
	)
}

export const EmptyState = ({ message, command }: { readonly message: string; readonly command?: string }) => (
	<div className='border-border text-dim space-y-3 border border-dashed px-4 py-6 text-sm'>
		<p>{message}</p>
		{command !== undefined && (
			<p className='flex flex-wrap items-center gap-2 text-xs'>
				Create one with
				<Command command={command} />
			</p>
		)}
	</div>
)
