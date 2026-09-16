type Props = {
	readonly title: string
	readonly children?: React.ReactNode
}

const RecordGone = ({ title, children }: Props) => (
	<div className='space-y-3 py-8'>
		<p className='text-dim text-sm'>
			{title ? <span className='text-foreground'>{title}</span> : 'This record'} was deleted.
		</p>

		{children}
	</div>
)

export { RecordGone }
