import type { HighlighterCore } from 'shiki/core'

// Only the languages folio bodies actually contain. Shiki's full bundle ships
// every grammar, which is megabytes; this list keeps the lazy chunk small.
const langs = {
	bash: () => import('shiki/langs/bash.mjs'),
	css: () => import('shiki/langs/css.mjs'),
	go: () => import('shiki/langs/go.mjs'),
	html: () => import('shiki/langs/html.mjs'),
	json: () => import('shiki/langs/json.mjs'),
	markdown: () => import('shiki/langs/markdown.mjs'),
	sql: () => import('shiki/langs/sql.mjs'),
	tsx: () => import('shiki/langs/tsx.mjs'),
	typescript: () => import('shiki/langs/typescript.mjs'),
	yaml: () => import('shiki/langs/yaml.mjs'),
} as const

export type Language = keyof typeof langs

const aliases: Record<string, Language> = {
	sh: 'bash',
	shell: 'bash',
	zsh: 'bash',
	js: 'typescript',
	javascript: 'typescript',
	ts: 'typescript',
	jsx: 'tsx',
	md: 'markdown',
	yml: 'yaml',
	golang: 'go',
}

export const resolveLanguage = (name: string | undefined): Language | undefined => {
	if (name === undefined) return undefined

	const lower = name.toLowerCase()

	if (lower in langs) return lower as Language

	return aliases[lower]
}

let pending: Promise<HighlighterCore> | undefined

const loaded = new Set<Language>()

// The highlighter is built once and shared. Creating one per block would
// reload every grammar for each fence on the page.
const core = async (): Promise<HighlighterCore> => {
	if (pending === undefined) {
		pending = (async () => {
			const [{ createHighlighterCore }, { createJavaScriptRegexEngine }] = await Promise.all([
				import('shiki/core'),
				import('shiki/engine/javascript'),
			])

			return createHighlighterCore({
				themes: [import('shiki/themes/github-light-default.mjs'), import('shiki/themes/github-dark-default.mjs')],
				langs: [],
				engine: createJavaScriptRegexEngine(),
			})
		})()
	}

	return pending
}

// Grammars load per language rather than up front: a page showing one SQL
// fence should not also download Go, TypeScript and the rest.
export const getHighlighter = async (language: Language): Promise<HighlighterCore> => {
	const highlighter = await core()

	if (!loaded.has(language)) {
		await highlighter.loadLanguage(await langs[language]())
		loaded.add(language)
	}

	return highlighter
}

export const THEMES = { light: 'github-light-default', dark: 'github-dark-default' } as const
