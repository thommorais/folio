import { screen } from '@testing-library/react'
import type { IssueId } from '_/core/domain/issue'
import { aCycle, anIssue } from '_/test/records'
import { renderPage } from '_/test/render-page'
import { describe, expect, it } from 'vitest'
import { TicketBody } from './index'

const work = anIssue('work', { title: 'Make the work legible' })
const theMap = anIssue('map', { wayfinder: 'map', title: 'Plan the graph view', parentId: work.id })

describe('TicketBody', () => {
	it('says which cycle a map plans', async () => {
		renderPage(<TicketBody project='folio' ticket={theMap} />, {
			issues: [work, theMap],
			cycles: [aCycle('c2', { ticketId: work.id, ordinal: 2, mapId: theMap.id })],
		})

		expect(await screen.findByText(/plans cycle 2 of/)).toBeDefined()
		expect(screen.getByRole('link', { name: 'Make the work legible' })).toBeDefined()
	})

	it('falls back to its parent when it plans no cycle', async () => {
		renderPage(<TicketBody project='folio' ticket={{ ...theMap, id: 'loose' as IssueId }} />, {
			issues: [work, theMap],
		})

		expect(await screen.findByText(/under/)).toBeDefined()
		expect(screen.queryByText(/plans cycle/)).toBeNull()
	})
})
