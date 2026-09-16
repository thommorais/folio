import { useEffect, useRef } from 'react'

type Addressed = { readonly id: string; readonly slug: string }

type Options = {
	readonly current: string
	readonly record: Addressed | undefined
	readonly rename: (slug: string) => void
}

export const useSlugSync = ({ current, record, rename }: Options): void => {
	const addressed = useRef<{ readonly url: string; readonly id: string } | undefined>(undefined)

	if (record?.slug === current) addressed.current = { url: current, id: record.id }

	const slug = record?.slug
	const id = record?.id

	useEffect(() => {
		if (!slug || slug === current) return

		const claim = addressed.current
		if (!claim || claim.url !== current || claim.id !== id) return

		rename(slug)
		// oxlint-disable-next-line react-hooks/exhaustive-deps
	}, [slug, id, current])
}
