import { Check, Copy } from 'lucide-react'
import { useEffect, useState } from 'react'
import { getHighlighter, resolveLanguage, THEMES } from './highlighter'

const CopyButton = ({ code }: { readonly code: string }) => {
	const [copied, setCopied] = useState(false)

	useEffect(() => {
		if (!copied) return

		const timer = setTimeout(() => setCopied(false), 2000)

		return () => clearTimeout(timer)
	}, [copied])

	return (
		<button
			type='button'
			aria-label={copied ? 'Copied' : 'Copy code'}
			onClick={() => {
				void navigator.clipboard.writeText(code).then(() => setCopied(true))
			}}
			className='text-dim hover:text-foreground absolute top-2 right-2 cursor-pointer p-1 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100'
		>
			{copied ? <Check size={14} /> : <Copy size={14} />}
		</button>
	)
}

const Frame = ({ label, children, code }: { label?: string; children: React.ReactNode; code: string }) => (
	<div className='group border-border relative border'>
		{label !== undefined && (
			<div className='border-border text-dim flex h-9 items-center border-b px-4 font-mono text-xs'>{label}</div>
		)}
		<CopyButton code={code} />
		{children}
	</div>
)

type Props = {
	readonly code: string
	readonly language: string | undefined
}

export const CodeBlock = ({ code, language }: Props) => {
	const resolved = resolveLanguage(language)
	const [html, setHtml] = useState<string>()

	useEffect(() => {
		if (resolved === undefined) return

		let cancelled = false

		void getHighlighter(resolved).then(highlighter => {
			if (cancelled) return

			setHtml(highlighter.codeToHtml(code, { lang: resolved, themes: THEMES, defaultColor: false }))
		})

		return () => {
			cancelled = true
		}
	}, [code, resolved])

	// Until the grammar loads, and for a fence with no language, the code still
	// reads as plain monospace rather than disappearing.
	if (html === undefined) {
		return (
			<Frame label={language} code={code}>
				<pre className='overflow-x-auto p-4 text-xs leading-relaxed'>
					<code>{code}</code>
				</pre>
			</Frame>
		)
	}

	return (
		<Frame label={language} code={code}>
			{/* Shiki returns a <pre> of spans it has already escaped. */}
			<div className='shiki-host overflow-x-auto text-xs leading-relaxed' dangerouslySetInnerHTML={{ __html: html }} />
		</Frame>
	)
}
