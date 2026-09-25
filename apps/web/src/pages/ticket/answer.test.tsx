import { screen } from '@testing-library/react'
import type { EntryId } from '_/core/domain/entry'
import type { IssueId } from '_/core/domain/issue'
import { anEntry, anIssue } from '_/test/records'
import { renderPage } from '_/test/render-page'
import { describe, expect, it } from 'vitest'
import { Answer } from './answer'

describe('Answer', () => {
	const ticket = anIssue('tree-or-graph', {
		status: 'done',
		wayfinder: 'grilling',
		resolution: 'A graph.',
		resolutionEntry: 'why' as EntryId,
	})

	it('states the answer and renders its linked detail', async () => {
		renderPage(<Answer project='folio' ticket={ticket} />, {
			entries: [
				anEntry('why', { kind: 'resolution', issueId: ticket.id, body: 'The tree hides blockers.' }),
				anEntry('other', { kind: 'resolution', issueId: ticket.id, body: 'An older draft.' }),
			],
		})

		expect(await screen.findByText('A graph.')).toBeDefined()
		expect(await screen.findByText('The tree hides blockers.')).toBeDefined()
		expect(screen.queryByText('An older draft.')).toBeNull()
	})

	it('states the answer alone when no detail is linked', async () => {
		renderPage(<Answer project='folio' ticket={{ ...ticket, resolutionEntry: undefined }} />, {
			entries: [anEntry('stray', { kind: 'resolution', issueId: 'x' as IssueId, body: 'Not this one.' })],
		})

		expect(await screen.findByText('A graph.')).toBeDefined()
		expect(screen.queryByText('Not this one.')).toBeNull()
	})
})
