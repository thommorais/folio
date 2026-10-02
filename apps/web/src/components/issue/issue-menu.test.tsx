import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ContainerProvider, type Container } from '_/app/container'
import { projectId } from '_/core/domain/project'
import { err, ok } from '_/lib/result'
import { anIssue } from '_/test/records'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { IssueMenu } from './issue-menu'

const toast = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn() }))
vi.mock('@thom/ui/toast', () => ({ toast }))

afterEach(() => {
	cleanup()
	vi.resetAllMocks()
})

const issue = anIssue('i1', { title: 'Fix it' })

const mount = (remove: ReturnType<typeof vi.fn>, props: { readonly share?: boolean } = {}) => {
	const onDeleted = vi.fn()
	const container = { issues: { remove, update: vi.fn() } } as unknown as Container

	render(
		<ContainerProvider container={container}>
			<IssueMenu
				issue={issue}
				onDeleted={onDeleted}
				onArchived={() => {}}
				share={props.share ? { kind: 'issue', id: 'i1', projectId: projectId('p1') } : undefined}
			/>
		</ContainerProvider>,
	)

	return { onDeleted, user: userEvent.setup() }
}

const openMenu = (user: ReturnType<typeof userEvent.setup>) =>
	user.click(screen.getByRole('button', { name: 'More actions' }))

describe('IssueMenu', () => {
	it('lists Edit and Delete', async () => {
		const { user } = mount(vi.fn())

		await openMenu(user)

		expect(screen.getByRole('menuitem', { name: 'Edit' })).toBeDefined()
		expect(screen.getByRole('menuitem', { name: 'Delete' })).toBeDefined()
		expect(screen.queryByRole('menuitem', { name: 'Share' })).toBeNull()
	})

	it('lists Share only when given a share target', async () => {
		const { user } = mount(vi.fn(), { share: true })

		await openMenu(user)

		expect(screen.getByRole('menuitem', { name: 'Share' })).toBeDefined()
	})

	it('asks before deleting and does not delete on the first click', async () => {
		const remove = vi.fn()
		const { user } = mount(remove)

		await openMenu(user)
		await user.click(screen.getByRole('menuitem', { name: 'Delete' }))

		expect(screen.getByText('Delete this issue?')).toBeDefined()
		expect(remove).not.toHaveBeenCalled()
	})

	it('goes back to the menu on Keep', async () => {
		const remove = vi.fn()
		const { user } = mount(remove)

		await openMenu(user)
		await user.click(screen.getByRole('menuitem', { name: 'Delete' }))
		await user.click(screen.getByRole('button', { name: 'Keep' }))

		expect(screen.getByRole('menuitem', { name: 'Delete' })).toBeDefined()
		expect(remove).not.toHaveBeenCalled()
	})

	it('deletes the issue on confirm, then reports it', async () => {
		const remove = vi.fn(async () => ok(undefined))
		const { user, onDeleted } = mount(remove)

		await openMenu(user)
		await user.click(screen.getByRole('menuitem', { name: 'Delete' }))
		await user.click(screen.getByRole('button', { name: 'Delete' }))

		await waitFor(() => expect(onDeleted).toHaveBeenCalledTimes(1))
		expect(remove).toHaveBeenCalledWith('i1')
		expect(toast.success).toHaveBeenCalledWith('Deleted.')
	})

	it('shows the error and does not report a deletion when it fails', async () => {
		const remove = vi.fn(async () => err(new Error('Could not delete the issue: forbidden')))
		const { user, onDeleted } = mount(remove)

		await openMenu(user)
		await user.click(screen.getByRole('menuitem', { name: 'Delete' }))
		await user.click(screen.getByRole('button', { name: 'Delete' }))

		await waitFor(() => expect(toast.error).toHaveBeenCalledWith('Could not delete the issue: forbidden'))
		expect(onDeleted).not.toHaveBeenCalled()
		expect(toast.success).not.toHaveBeenCalled()
	})

	it('opens the edit sheet from Edit', async () => {
		const { user } = mount(vi.fn())

		await openMenu(user)
		await user.click(screen.getByRole('menuitem', { name: 'Edit' }))

		expect(await screen.findByText('Edit ticket')).toBeDefined()
		expect((screen.getByLabelText('Title') as HTMLInputElement).value).toBe('Fix it')
	})
})
