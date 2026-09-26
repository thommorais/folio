import { beforeEach, describe, expect, it } from 'vitest'
import { createLastProject } from './last-project'

const scope = { client: 'acme', domain: 'web', slug: 'site' }

describe('lastProject', () => {
	beforeEach(() => localStorage.clear())

	it('resumes the project remembered in an earlier session', () => {
		createLastProject().remember(scope)

		expect(createLastProject().resume()).toEqual(scope)
	})

	it('resumes only once per session, so home stays reachable', () => {
		createLastProject().remember(scope)
		const session = createLastProject()

		session.resume()

		expect(session.resume()).toBeUndefined()
	})

	it('stops resuming once the first page has rendered, wherever the session started', () => {
		createLastProject().remember(scope)
		const session = createLastProject()

		session.settle()

		expect(session.resume()).toBeUndefined()
	})

	it('has nothing to resume before any project was opened', () => {
		expect(createLastProject().resume()).toBeUndefined()
	})

	it('ignores a stored value that is not a project', () => {
		localStorage.setItem('folio.last-project', '{"client":1}')

		expect(createLastProject().resume()).toBeUndefined()
	})
})
