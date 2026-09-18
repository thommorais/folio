import { createFileRoute } from '@tanstack/react-router'
import { JournalEntryDetail } from '_/pages/journal-entry'

export const Route = createFileRoute('/_authenticated/$client/$domain/$slug/journal/$entry')({
	component: JournalEntryDetail,
})
