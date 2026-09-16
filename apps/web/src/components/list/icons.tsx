const SearchIcon = () => (
	<svg viewBox='0 0 16 16' fill='none' className='size-4' aria-hidden>
		<circle cx='7' cy='7' r='4.25' stroke='currentColor' strokeWidth='1.5' />
		<path d='M10.5 10.5L14 14' stroke='currentColor' strokeWidth='1.5' strokeLinecap='round' />
	</svg>
)

const FilterIcon = () => (
	<svg viewBox='0 0 16 16' fill='none' className='size-4' aria-hidden>
		<path d='M2 4h12M4 8h8M6.5 12h3' stroke='currentColor' strokeWidth='1.5' strokeLinecap='round' />
	</svg>
)

const ClearIcon = ({ className }: { readonly className?: string }) => (
	<svg viewBox='0 0 16 16' fill='none' className={className} aria-hidden>
		<path d='M4 4l8 8M12 4l-8 8' stroke='currentColor' strokeWidth='1.5' strokeLinecap='round' />
	</svg>
)

const SortIcon = ({ direction }: { readonly direction: 'asc' | 'desc' | undefined }) => (
	<svg viewBox='0 0 16 16' fill='none' className='size-4' aria-hidden>
		<path d='M4 6h8M4 10h5' stroke='currentColor' strokeWidth='1.5' strokeLinecap='round' />
		{direction !== undefined && (
			<path
				d={direction === 'asc' ? 'M12.5 12V8M11 9.5l1.5-1.5L14 9.5' : 'M12.5 8v4M11 10.5l1.5 1.5L14 10.5'}
				stroke='currentColor'
				strokeWidth='1.5'
				strokeLinecap='round'
				strokeLinejoin='round'
			/>
		)}
	</svg>
)

export { ClearIcon, FilterIcon, SearchIcon, SortIcon }
