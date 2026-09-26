import { cn } from '@thom/libs/cn'
import type { ReactNode } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

const Inline = ({ children }: { readonly children?: ReactNode }) => <span>{children} </span>

const Skip = () => null

export const MarkdownPreview = ({ children, className }: { readonly children: string; readonly className?: string }) => (
	<div className={cn('text-dim line-clamp-2 text-sm', className)}>
		<ReactMarkdown
			remarkPlugins={[remarkGfm]}
			components={{
				p: Inline,
				h1: Inline,
				h2: Inline,
				h3: Inline,
				h4: Inline,
				h5: Inline,
				h6: Inline,
				ul: Inline,
				ol: Inline,
				li: Inline,
				blockquote: Inline,
				a: Inline,
				pre: ({ children }) => <>{children}</>,
				code: ({ children }) => <code className='font-mono text-xs'>{children}</code>,
				strong: ({ children }) => <strong className='text-foreground font-medium'>{children}</strong>,
				input: Skip,
				img: Skip,
				table: Skip,
				hr: Skip,
				br: () => ' ',
			}}
		>
			{children}
		</ReactMarkdown>
	</div>
)
