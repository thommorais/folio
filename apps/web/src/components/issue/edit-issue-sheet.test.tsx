import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ContainerProvider, type Container } from '_/app/container'
import { err, ok } from '_/lib/result'
import { anIssue } from '_/test/records'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { EditIssueSheet } from './edit-issue-sheet'

const toast = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn() }))
vi.mock('@thom/ui/toast', () => ({ toast }))

afterEach(() => {
	cleanup()
	vi.resetAllMocks()
})

const open = (update: ReturnType<typeof vi.fn>, issue = anIssue('i1', { title: 'Old', body: 'Body', priority: 'low' })) => {
	const onOpenChange = vi.fn()
	const container = { issues: { update } } as unknown as Container

	render(
		<ContainerProvider container={container}>
			<EditIssueSheet issue={issue} open onOpenChange={onOpenChange} />
		</ContainerProvider>,
	)

	return { onOpenChange }
}

const field = (label: string) => screen.getByLabelText(label) as HTMLInputElement

describe('EditIssueSheet', () => {
	it('starts from the current values', () => {
		open(vi.fn(), anIssue('i1', { title: 'Old', body: 'Body', priority: 'high', size: 3 }))

		expect(field('Title').value).toBe('Old')
		expect(field('Description').value).toBe('Body')
		expect(field('Priority').value).toBe('high')
		expect(field('Size').value).toBe('3')
	})

	it('starts unsized when the issue has no size', () => {
		open(vi.fn())

		expect(field('Size').value).toBe('')
	})

	it('titles the sheet by kind', () => {
		open(vi.fn(), anIssue('i1', { kind: 'todo' }))

		expect(screen.getByText('Edit todo')).toBeDefined()
	})

	it('saves the trimmed title with the other fields and closes', async () => {
		const update = vi.fn(async () => ok(undefined))
		const { onOpenChange } = open(update)
		const user = userEvent.setup()

		await user.clear(field('Title'))
		await user.type(field('Title'), '  New  ')
		await user.selectOptions(field('Priority'), 'high')
		await user.selectOptions(field('Size'), '5')
		await user.click(screen.getByRole('button', { name: 'Save' }))

		await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false))
		expect(update).toHaveBeenCalledWith('i1', { title: 'New', body: 'Body', priority: 'high', size: 5 })
		expect(toast.success).toHaveBeenCalledWith('Saved.')
	})

	it('sends a null size when Unsized is chosen', async () => {
		const update = vi.fn(async () => ok(undefined))
		open(update, anIssue('i1', { size: 8 }))
		const user = userEvent.setup()

		await user.selectOptions(field('Size'), '')
		await user.click(screen.getByRole('button', { name: 'Save' }))

		await waitFor(() => expect(update).toHaveBeenCalled())
		expect(update).toHaveBeenCalledWith('i1', expect.objectContaining({ size: null }))
	})

	it('keeps Save disabled while the title is blank', async () => {
		open(vi.fn())
		const user = userEvent.setup()

		await user.clear(field('Title'))

		expect((screen.getByRole('button', { name: 'Save' }) as HTMLButtonElement).disabled).toBe(true)
	})

	it('shows the error and stays open when the save fails', async () => {
		const update = vi.fn(async () => err(new Error('Could not save the issue: slug taken')))
		const { onOpenChange } = open(update)
		const user = userEvent.setup()

		await user.click(screen.getByRole('button', { name: 'Save' }))

		expect(await screen.findByText('Could not save the issue: slug taken')).toBeDefined()
		expect(onOpenChange).not.toHaveBeenCalled()
		expect(toast.success).not.toHaveBeenCalled()
	})
})
