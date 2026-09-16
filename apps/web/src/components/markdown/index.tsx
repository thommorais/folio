import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { CodeBlock } from './code-block'

const languageOf = (className: string | undefined): string | undefined =>
	className
		?.split(' ')
		.find(name => name.startsWith('language-'))
		?.replace('language-', '')

export const Markdown = ({ children }: { readonly children: string }) => (
	<div className='space-y-4 text-sm leading-relaxed'>
		<ReactMarkdown
			remarkPlugins={[remarkGfm]}
			components={{
				// react-markdown gives a fenced block as <pre><code>. The fence is
				// unwrapped here so CodeBlock owns the frame instead of nesting
				// inside a <pre> it cannot style.
				pre: ({ children }) => <>{children}</>,
				code: ({ className, children, ...props }) => {
					const text = String(children).replace(/\n$/, '')
					const language = languageOf(className)

					if (language === undefined && !text.includes('\n')) {
						return (
							<code className='bg-accent border-border border px-1 py-0.5 font-mono text-xs' {...props}>
								{children}
							</code>
						)
					}

					return <CodeBlock code={text} language={language} />
				},
				h1: ({ children }) => <h2 className='pt-2 font-serif text-lg'>{children}</h2>,
				h2: ({ children }) => <h2 className='pt-2 font-serif text-lg'>{children}</h2>,
				h3: ({ children }) => <h3 className='pt-2 text-sm font-medium'>{children}</h3>,
				p: ({ children }) => <p>{children}</p>,
				ul: ({ children }) => <ul className='list-disc space-y-1 pl-5'>{children}</ul>,
				ol: ({ children }) => <ol className='list-decimal space-y-1 pl-5'>{children}</ol>,
				blockquote: ({ children }) => (
					<blockquote className='border-border text-dim border-l-2 pl-4'>{children}</blockquote>
				),
				a: ({ children, href }) => (
					<a
						href={href}
						target='_blank'
						rel='noreferrer'
						className='hover:text-foreground underline underline-offset-2 transition-colors'
					>
						{children}
					</a>
				),
				table: ({ children }) => (
					<div className='overflow-x-auto'>
						<table className='border-border w-full border text-xs'>{children}</table>
					</div>
				),
				th: ({ children }) => <th className='border-border border px-3 py-2 text-left font-medium'>{children}</th>,
				td: ({ children }) => <td className='border-border border px-3 py-2'>{children}</td>,
				hr: () => <hr className='border-border' />,
			}}
		>
			{children}
		</ReactMarkdown>
	</div>
)
