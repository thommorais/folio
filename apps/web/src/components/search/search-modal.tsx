import { CommandDialog, CommandDialogContent, CommandDialogTitle } from '@thom/ui/command';
import { useSearchStore } from '_/app/search-store';
import { useHotkeys } from 'react-hotkeys-hook';
import { Search } from './search';
import { SearchFooter } from './search-footer';

export const SearchModal = () => {
	const { isOpen, setOpen } = useSearchStore()

	useHotkeys('meta+k', () => setOpen(), { enableOnFormTags: true })

	return (
		<CommandDialog open={isOpen} onOpenChange={setOpen}>
			<CommandDialogContent className='m-0 h-130 w-[96dvw] max-w-full border-none bg-transparent p-0 md:max-w-185'>
				<CommandDialogTitle className='sr-only'>Search</CommandDialogTitle>

				{isOpen && (
					<>
						<Search />
						<SearchFooter />
					</>
				)}
			</CommandDialogContent>
		</CommandDialog>
	)
}
