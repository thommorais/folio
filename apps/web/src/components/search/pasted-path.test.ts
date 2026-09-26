import { describe, expect, it } from 'vitest'
import { pastedPath } from './pasted-path'

const origin = 'https://folio.journ.app'

describe('pastedPath', () => {
	it('opens a pasted link to a page in this app', () => {
		expect(pastedPath('https://folio.journ.app/acme/web/site/tickets/fix-nav', origin)).toBe('/acme/web/site/tickets/fix-nav')
	})

	it('keeps the query string', () => {
		expect(pastedPath('https://folio.journ.app/acme/web/site/todos?q=nav', origin)).toBe('/acme/web/site/todos?q=nav')
	})

	it('tolerates surrounding whitespace', () => {
		expect(pastedPath('  https://folio.journ.app/knowledge/tip \n', origin)).toBe('/knowledge/tip')
	})

	it('ignores links to other sites', () => {
		expect(pastedPath('https://example.com/acme/web/site', origin)).toBeUndefined()
	})

	it('ignores the bare home link, which has nothing to jump to', () => {
		expect(pastedPath('https://folio.journ.app/', origin)).toBeUndefined()
	})

	it('treats ordinary text as a search', () => {
		expect(pastedPath('search ranking', origin)).toBeUndefined()
	})
})
