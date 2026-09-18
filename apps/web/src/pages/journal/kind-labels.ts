import type { EntryKind } from '_/core/domain/entry'

export const ENTRY_KIND_LABELS: Record<EntryKind, string> = {
	journal: 'Journal',
	doc: 'Doc',
	log: 'Log',
}
