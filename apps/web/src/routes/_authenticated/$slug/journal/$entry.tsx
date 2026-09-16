import { createFileRoute } from '@tanstack/react-router'
import { JournalEntryDetail } from '_/pages/journal-entry'

export const Route = createFileRoute('/_authenticated/$slug/journal/$entry')({
	component: JournalEntryDetail,
})
