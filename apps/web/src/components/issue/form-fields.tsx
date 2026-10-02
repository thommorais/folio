export const controlClasses =
	'border-border text-foreground placeholder:text-muted-foreground focus-visible:border-ring w-full border bg-transparent px-3 text-sm transition-colors focus-visible:outline-none'

export const FieldLabel = ({ htmlFor, children }: { readonly htmlFor: string; readonly children: React.ReactNode }) => (
	<label htmlFor={htmlFor} className='block text-sm font-medium'>
		{children}
	</label>
)

export const Hint = ({ children }: { readonly children: React.ReactNode }) => <p className='text-dim text-xs'>{children}</p>
