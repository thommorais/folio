import { screen, within } from '@testing-library/react'
import type { IssueId } from '_/core/domain/issue'
import { anIssue } from '_/test/records'
import { renderPage } from '_/test/render-page'
import { describe, expect, it } from 'vitest'
import { MapFrontier } from './map-frontier'

const theMap = anIssue('map', { wayfinder: 'map', title: 'Plan the view' })
const under = { parentId: theMap.id }

const decided = anIssue('shape', {
	...under,
	wayfinder: 'grilling',
	title: 'Decide the shape',
	status: 'done',
	resolution: 'A graph.',
})
const dropped = anIssue('lib', {
	...under,
	wayfinder: 'task',
	title: 'Embed a library',
	status: 'cancelled',
	resolution: 'Past the destination.',
})
const waitsOnDone = anIssue('rail', {
	...under,
	wayfinder: 'prototype',
	title: 'Prototype the rail',
	dependsOn: [decided.id],
})
const open = anIssue('dense', {
	...under,
	wayfinder: 'research',
	title: 'How dense',
	status: 'in_progress',
	assignee: 'u1' as never,
})
const waitsOnOpen = anIssue('edges', {
	...under,
	wayfinder: 'task',
	title: 'Draw the edges',
	dependsOn: [open.id],
	blocked: true,
})

const section = (name: RegExp) => screen.getByRole('heading', { name }).parentElement as HTMLElement

describe('MapFrontier', () => {
	it('lists each closed decision with its answer, and ruled-out work apart from it', async () => {
		renderPage(<MapFrontier project='folio' map={theMap} />, { issues: [theMap, decided, dropped, waitsOnDone] })

		await screen.findByRole('heading', { name: /Decisions so far/ })

		const decisions = section(/Decisions so far/)
		expect(within(decisions).getByText('Decide the shape')).toBeDefined()
		expect(within(decisions).getByText('A graph.')).toBeDefined()
		expect(within(decisions).queryByText('Embed a library')).toBeNull()

		const outOfScope = section(/Out of scope/)
		expect(within(outOfScope).getByText('Past the destination.')).toBeDefined()
	})

	it('counts only the blockers still open', async () => {
		renderPage(<MapFrontier project='folio' map={theMap} />, {
			issues: [theMap, decided, open, waitsOnDone, waitsOnOpen],
		})

		await screen.findByRole('heading', { name: /Takeable/ })

		expect(within(section(/Takeable/)).queryByText(/waits on/)).toBeNull()
		expect(within(section(/Blocked/)).getByText('waits on 1')).toBeDefined()
	})

	it('says so when the map has no decisions yet', async () => {
		renderPage(<MapFrontier project='folio' map={{ ...theMap, id: 'empty' as IssueId }} />, { issues: [theMap] })

		expect(await screen.findByText('No decision tickets on this map yet.')).toBeDefined()
	})
})
