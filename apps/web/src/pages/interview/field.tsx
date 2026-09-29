type Props = {
	readonly value: string
	readonly placeholder: string
	readonly disabled: boolean
	readonly onChange: (value: string) => void
}

export const Field = ({ value, placeholder, disabled, onChange }: Props) => (
	<textarea
		rows={3}
		value={value}
		disabled={disabled}
		placeholder={placeholder}
		onChange={event => onChange(event.target.value)}
		className='border-border placeholder:text-dimmer focus:border-foreground/40 block w-full resize-y border bg-transparent px-3 py-2 text-sm transition-colors outline-none disabled:opacity-60'
	/>
)
