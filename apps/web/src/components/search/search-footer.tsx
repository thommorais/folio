import { ArrowDown, ArrowUp, CornerDownLeft } from 'lucide-react'

const keys = [ArrowUp, ArrowDown, CornerDownLeft]

export const SearchFooter = () => (
	<div className='search-footer bg-background border-border flex h-[40px] w-full items-center border border-t-0 px-3 backdrop-blur-lg dark:bg-[#0C0C0C]/99'>
		<span className='text-dimmer font-serif text-sm'>folio</span>

		<div className='ml-auto flex space-x-2'>
			{keys.map(Key => (
				<div
					key={Key.displayName}
					className='bg-accent border-border flex size-6 items-center justify-center border select-none'
				>
					<Key className='text-foreground size-3' />
				</div>
			))}
		</div>
	</div>
)
