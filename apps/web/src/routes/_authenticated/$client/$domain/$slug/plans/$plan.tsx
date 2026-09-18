import { createFileRoute } from '@tanstack/react-router'
import { PlanDetail } from '_/pages/plan'

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/plans/$plan')({
	component: PlanDetail,
})
