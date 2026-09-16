import { PreviewSheet } from './preview-sheet'
import { OpenSearchButton } from './search/open-search-button'
import { SearchModal } from './search/search-modal'
import { UserMenu } from './user-menu'

const Header = () => (
	<header className='border-border group flex h-[70px] items-center justify-between border-b px-4 md:px-6'>
		<OpenSearchButton />

		<div className='ml-auto flex items-center space-x-2'>
			<UserMenu />
		</div>
	</header>
)

export const AppShell = ({ children }: { children: React.ReactNode }) => (
	<div className='bg-background relative flex min-h-dvh w-full flex-col'>
		<Header />
		<main className='min-w-0 flex-1 px-4 py-8 md:px-6'>{children}</main>

		<SearchModal />
		<PreviewSheet />
	</div>
)
