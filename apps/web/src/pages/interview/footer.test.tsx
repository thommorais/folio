import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Footer } from './footer'

afterEach(cleanup)

const props = {
	staged: {},
	count: 0,
	canSend: true,
	canFinish: false,
	sending: false,
	pending: false,
	working: false,
	locked: false,
	error: undefined,
	onSend: () => {},
}

const button = (name: string | RegExp) => screen.getByRole('button', { name }) as HTMLButtonElement

describe('Footer', () => {
	it('says nothing is staged and keeps Send disabled', () => {
		render(<Footer {...props} />)

		expect(screen.getByText(/Nothing staged/)).toBeDefined()
		expect(button('Send').disabled).toBe(true)
	})

	it('summarises what is staged and puts the count on Send', () => {
		render(
			<Footer {...props} count={3} staged={{ q1: { answer: { kind: 'accept', option: 'a' }, thread: ['x', 'y'] } }} />,
		)

		expect(screen.getByText('Staged: 1 answer, 2 messages.')).toBeDefined()
		expect(button('Send 3').disabled).toBe(false)
	})

	it('tells the user to continue the agent after a Send', () => {
		render(<Footer {...props} pending canSend={false} />)

		expect(screen.getByText('Sent. Tell the agent to continue.')).toBeDefined()
		expect(button('Send').disabled).toBe(true)
	})

	it('says the interview is finished, ahead of any pending state, and disables both buttons', () => {
		render(<Footer {...props} locked pending canSend={false} canFinish={false} />)

		expect(screen.getByText('This interview is finished.')).toBeDefined()
		expect(button('Finish').disabled).toBe(true)
		expect(button('Send').disabled).toBe(true)
	})

	it('shows a failed Send in place of the status', () => {
		render(<Footer {...props} count={1} error='Q1.answer: must name one of the options.' />)

		expect(screen.getByText('Q1.answer: must name one of the options.')).toBeDefined()
	})

	it('asks for confirmation before Finish and sends it with the finish flag', async () => {
		const onSend = vi.fn()
		render(<Footer {...props} canFinish onSend={onSend} />)

		await userEvent.click(button('Finish'))
		expect(onSend).not.toHaveBeenCalled()
		expect(screen.getByText(/Close the interview after this Send/)).toBeDefined()

		await userEvent.click(button('Finish'))

		expect(onSend).toHaveBeenCalledWith(true)
	})

	it('cancels the Finish confirmation without sending', async () => {
		const onSend = vi.fn()
		render(<Footer {...props} canFinish onSend={onSend} />)

		await userEvent.click(button('Finish'))
		await userEvent.click(button('Cancel'))

		expect(onSend).not.toHaveBeenCalled()
		expect(button('Finish')).toBeDefined()
	})
})
