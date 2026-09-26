import { describe, expect, it } from 'vitest'
import { collapseCrumbs } from './collapse-crumbs'

const crumb = (key: string, title: string) => ({ key, title })

describe('collapseCrumbs', () => {
	it('keeps the deepest of a run of identical names', () => {
		const crumbs = [crumb('root', 'Clients'), crumb('client', 'folio'), crumb('domain', 'folio'), crumb('project', 'folio'), crumb('section', 'Journal')]

		expect(collapseCrumbs(crumbs).map(entry => entry.key)).toEqual(['root', 'project', 'section'])
	})

	it('treats names that differ only in case as the same', () => {
		expect(collapseCrumbs([crumb('client', 'Folio'), crumb('domain', 'folio')]).map(entry => entry.key)).toEqual(['domain'])
	})

	it('leaves distinct names alone', () => {
		const crumbs = [crumb('client', 'acme'), crumb('domain', 'web'), crumb('project', 'acme')]

		expect(collapseCrumbs(crumbs)).toEqual(crumbs)
	})
})
