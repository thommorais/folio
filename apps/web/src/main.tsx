import './styles/globals.css'
import './styles/tailwind.css'

import { createRouter, RouterProvider } from '@tanstack/react-router'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { routeTree } from './routeTree.gen'
import { listMemory } from './adapters/browser/list-memory'

const router = createRouter({ routeTree })

const LISTS: ReadonlySet<string> = new Set([
	'/_authenticated/$client/$domain/$slug/tickets/',
	'/_authenticated/$client/$domain/$slug/todos/',
	'/_authenticated/$client/$domain/$slug/plans/',
	'/_authenticated/$client/$domain/$slug/journal/',
	'/_authenticated/$client/$domain/$slug/work',
])

router.subscribe('onResolved', () => {
	const leaf = router.state.matches.at(-1)
	if (leaf && LISTS.has(leaf.routeId)) listMemory.remember(leaf.routeId, leaf.search as Record<string, unknown>)
})

declare module '@tanstack/react-router' {
	interface Register {
		router: typeof router
	}
}

const rootElement = document.getElementById('root')
if (!rootElement) {
	throw new Error('Root element #root not found')
}

createRoot(rootElement).render(
	<StrictMode>
		<RouterProvider router={router} />
	</StrictMode>,
)
