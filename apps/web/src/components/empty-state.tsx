export const EmptyState = ({ message, command }: { readonly message: string; readonly command?: string }) => (
	<div className='border-border text-dim space-y-3 border border-dashed px-4 py-6 text-sm'>
		<p>{message}</p>
		{command !== undefined && (
			<p className='flex flex-wrap items-center gap-2 text-xs'>
				Create one with
				<code className='bg-accent text-foreground px-1.5 py-0.5 font-mono select-all'>{command}</code>
			</p>
		)}
	</div>
)
