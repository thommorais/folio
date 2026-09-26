export const TAG_TONES = [
	'bg-sky-500',
	'bg-violet-500',
	'bg-emerald-500',
	'bg-amber-500',
	'bg-rose-500',
	'bg-teal-500',
	'bg-orange-500',
	'bg-indigo-500',
] as const

export type TagTone = (typeof TAG_TONES)[number]

export const tagTone = (tag: string): TagTone => {
	let hash = 0
	for (const char of tag.trim().toLowerCase()) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
	return TAG_TONES[hash % TAG_TONES.length] as TagTone
}
