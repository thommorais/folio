import { Link } from '@tanstack/react-router'
import { ConnectionStatus } from './connection-status'
import { PreviewSheet } from './preview-sheet'
import { OpenSearchButton } from './search/open-search-button'
import { SearchModal } from './search/search-modal'
import { UserMenu } from './user-menu'
import { useRowNavigation } from '_/components/list/row-navigation'

// Knowledge sits in the header rather than under a project, because it belongs
// to none: there is no breadcrumb that would ever lead to it.
const Header = () => (
	<header className='border-border group flex h-[70px] items-center justify-between border-b px-4 md:px-6'>
		<OpenSearchButton />

		<div className='ml-auto flex items-center space-x-4'>
			<ConnectionStatus />
			<Link
				to='/knowledge'
				className='text-dim hover:text-foreground text-sm transition-colors'
				activeProps={{ className: 'text-foreground' }}
			>
				Knowledge
			</Link>
			<UserMenu />
		</div>
	</header>
)

export const AppShell = ({ children }: { children: React.ReactNode }) => {
	useRowNavigation()

	return (
		<div className='bg-background relative flex min-h-dvh w-full flex-col'>
			<Header />
			<main className='min-w-0 flex-1 px-4 py-8 md:px-6'>{children}</main>

			<SearchModal />
			<PreviewSheet />
		</div>
	)
}
