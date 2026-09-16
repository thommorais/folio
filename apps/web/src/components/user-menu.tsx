import { useNavigate } from '@tanstack/react-router'
import { Avatar } from '@thom/ui/avatar'
import { Dropdown, DropdownButton, DropdownDivider, DropdownItem, DropdownMenu } from '@thom/ui/dropdown'
import { initialsOf } from '_/core/domain/session'
import { useCurrentUser, useSignOut } from '_/app/use-session'
import { ThemeSwitch } from './theme-switch'

export const UserMenu = () => {
	const user = useCurrentUser()
	const signOut = useSignOut()
	const navigate = useNavigate()

	if (!user) return null

	const onSignOut = () => {
		signOut()
		void navigate({ to: '/login', replace: true })
	}

	return (
		<Dropdown>
			<DropdownButton as='button' className='cursor-pointer' aria-label='Account menu'>
				<Avatar src={user.avatarUrl} initials={initialsOf(user)} alt={user.name} className='size-8' />
			</DropdownButton>

			<DropdownMenu anchor='bottom end' className='w-[240px]'>
				<div className='flex flex-col px-3 py-2'>
					<span className='text-foreground truncate text-xs'>{user.name || user.email}</span>
					<span className='text-dimmer truncate text-xs'>{user.email}</span>
				</div>

				<DropdownDivider />

				<div className='flex items-center justify-between px-3 py-1.5'>
					<span className='text-xs'>Theme</span>
					<ThemeSwitch />
				</div>

				<DropdownDivider />

				<DropdownItem onClick={onSignOut} className='text-xs'>
					Sign out
				</DropdownItem>
			</DropdownMenu>
		</Dropdown>
	)
}
