export const pastedPath = (text: string, origin: string): string | undefined => {
	let url: URL
	try {
		url = new URL(text.trim())
	} catch {
		return undefined
	}

	if (url.origin !== origin || url.pathname === '/') return undefined

	return `${url.pathname}${url.search}`
}
