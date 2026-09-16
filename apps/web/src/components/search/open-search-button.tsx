import { Search } from 'lucide-react'
import { useSearchStore } from '_/app/search-store'

export const OpenSearchButton = () => {
	const setOpen = useSearchStore(state => state.setOpen)

	return (
		<button
			type='button'
			onClick={() => setOpen(true)}
			className='group/search text-muted-foreground hover:text-foreground relative flex w-full cursor-pointer items-center justify-start gap-2 text-sm font-normal transition-colors md:w-40 md:min-w-[250px] lg:w-64'
		>
			<Search size={18} className='transition-transform duration-200 group-hover/search:scale-110' />
			<span>Find anything...</span>
			<kbd className='bg-accent border-border pointer-events-none absolute top-0 right-0 hidden h-5 items-center gap-1 border px-1.5 text-[10px] font-medium opacity-0 transition-opacity duration-200 select-none group-hover/search:opacity-100 sm:flex'>
				<span className='text-xs'>⌘</span>K
			</kbd>
		</button>
	)
}
