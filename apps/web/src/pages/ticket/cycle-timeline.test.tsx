import { screen } from '@testing-library/react'
import type { IssueId } from '_/core/domain/issue'
import { aCycle, anIssue } from '_/test/records'
import { renderPage } from '_/test/render-page'
import { describe, expect, it } from 'vitest'
import { CycleTimeline } from './cycle-timeline'

const theMap = anIssue('map', { wayfinder: 'map', title: 'Plan the graph view', parentId: 'work' as IssueId })
const under = { parentId: theMap.id }
const decisions = [
	anIssue('shape', { ...under, wayfinder: 'grilling', status: 'done', resolution: 'A graph.' }),
	anIssue('dense', { ...under, wayfinder: 'research', status: 'in_progress' }),
	anIssue('rail', { ...under, wayfinder: 'prototype' }),
	anIssue('step', { ...under, kind: 'todo' }),
]

describe('CycleTimeline', () => {
	it('names the map planning a cycle, what is decided and what still holds plan', async () => {
		renderPage(<CycleTimeline project='folio' cycles={[aCycle('c1', { mapId: theMap.id })]} maps={[theMap]} />, {
			issues: [theMap, ...decisions],
		})

		expect(await screen.findByText('Plan the graph view')).toBeDefined()
		expect(screen.getByText('1 decided')).toBeDefined()
		expect(screen.getByText('2 open, plan holds')).toBeDefined()
	})

	it('drops the hold once the cycle has left plan', async () => {
		renderPage(
			<CycleTimeline project='folio' cycles={[aCycle('c1', { mapId: theMap.id, phase: 'do' })]} maps={[theMap]} />,
			{
				issues: [theMap, ...decisions],
			},
		)

		expect(await screen.findByText('Plan the graph view')).toBeDefined()
		expect(screen.queryByText(/plan holds/)).toBeNull()
	})

	it('shows no plan row for a cycle without a map', async () => {
		renderPage(
			<CycleTimeline project='folio' cycles={[aCycle('c1', { resolution: 'Shipped the list.' })]} maps={[theMap]} />,
			{
				issues: [theMap, ...decisions],
			},
		)

		expect(await screen.findByText('Shipped the list.')).toBeDefined()
		expect(screen.queryByText('Planned by')).toBeNull()
	})
})
