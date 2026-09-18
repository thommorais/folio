import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/')({
	beforeLoad: ({ params }) => {
		throw redirect({ to: '/$client/$domain/$slug/work', params })
	},
})
