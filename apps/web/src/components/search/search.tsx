import { useNavigate } from '@tanstack/react-router'
import { BookText, FolderKanban, ListTodo, NotebookPen, Search as SearchIcon } from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import { type ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { useHotkeys } from 'react-hotkeys-hook'
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@thom/ui/command'
import type { SearchHit, SearchKind } from '_/core/ports/search'
import { useSearch } from '_/app/use-search'
import { usePreviewStore } from '_/app/preview-store'
import { useSearchStore } from '_/app/search-store'

type Shortcut = {
	readonly id: string
	readonly title: string
	readonly action: () => void
}

const kindIcons: Record<SearchKind, typeof BookText> = {
	log: BookText,
	doc: NotebookPen,
	todo: ListTodo,
	plan: FolderKanban,
}

const groupLabels: Record<string, string> = {
	shortcut: 'Shortcuts',
	journal: 'Journal',
	doc: 'Docs',
	todo: 'Todos',
	plan: 'Plans',
}

const HitRow = ({ hit, index }: { hit: SearchHit; index: number }) => {
	const Icon = kindIcons[hit.kind]
	const reduced = useReducedMotion()

	const content = (
		<div className='flex w-full items-center gap-2'>
			<Icon className='text-dim size-4 shrink-0' />
			<span className='truncate'>{hit.title}</span>
			{hit.snippet && <span className='text-dim hidden truncate text-xs md:block'>{hit.snippet}</span>}
			<span className='text-dimmer ml-auto shrink-0 text-xs'>{hit.projectSlug}</span>
		</div>
	)

	if (reduced) return content

	return (
		<motion.div
			className='w-full'
			initial={{ opacity: 0, y: 4 }}
			animate={{ opacity: 1, y: 0 }}
			transition={{ duration: 0.18, ease: [0.16, 1, 0.3, 1], delay: Math.min(index, 6) * 0.02 }}
		>
			{content}
		</motion.div>
	)
}

export const Search = () => {
	const navigate = useNavigate()
	const setOpen = useSearchStore(state => state.setOpen)
	const openPreview = usePreviewStore(state => state.openPreview)
	const [term, setTerm] = useState('')
	const [debounced, setDebounced] = useState('')
	const wrapper = useRef<HTMLDivElement>(null)
	const list = useRef<HTMLDivElement>(null)

	// Multi-word queries are usually mid-typing, and each costs a request per project.
	const delay = term.trim().split(/\s+/u).length > 1 ? 700 : 200

	useEffect(() => {
		const timer = setTimeout(() => setDebounced(term), delay)
		return () => clearTimeout(timer)
	}, [term, delay])

	const { hits, isFetching } = useSearch(debounced)

	useHotkeys('esc', () => setOpen(false), { enableOnFormTags: true })

	const shortcuts: readonly Shortcut[] = useMemo(
		() => [
			{
				id: 'sc-view-projects',
				title: 'View projects',
				action: () => {
					setOpen(false)
					void navigate({ to: '/' })
				},
			},
		],
		[navigate, setOpen],
	)

	const openHit = (hit: SearchHit) => {
		setOpen(false)

		if (hit.kind === 'todo') {
			openPreview({ kind: 'todo', project: hit.projectSlug, id: hit.id })
			return
		}

		const project = { slug: hit.projectSlug }

		if (hit.kind === 'plan') {
			void navigate({ to: '/$slug/plans/$plan', params: { ...project, plan: hit.id } })
			return
		}

		if (hit.kind === 'doc') {
			void navigate({ to: '/$slug/docs/$doc', params: { ...project, doc: hit.slug } })
			return
		}

		void navigate({ to: '/$slug/journal/$entry', params: { ...project, entry: hit.slug } })
	}

	const grouped = useMemo(() => {
		const groups: Record<string, SearchHit[]> = {}

		for (const hit of hits) {
			const key = hit.kind
			groups[key] ??= []
			groups[key].push(hit)
		}

		const ordered: Record<string, SearchHit[]> = {}
		for (const key of ['log', 'doc', 'todo', 'plan']) {
			const group = groups[key]
			if (group && group.length > 0) ordered[key] = group
		}

		return ordered
	}, [hits])

	const matchingShortcuts = debounced
		? shortcuts.filter(shortcut => shortcut.title.toLowerCase().includes(debounced.toLowerCase()))
		: shortcuts

	// Feeds the list's natural height to CSS so the modal animates instead of jumping.
	useEffect(() => {
		if (!list.current || !wrapper.current) return

		const el = list.current
		const target = wrapper.current
		let frame: number

		const observer = new ResizeObserver(() => {
			frame = requestAnimationFrame(() => {
				target.style.setProperty('--search-list-height', `${el.offsetHeight}px`)
			})
		})

		observer.observe(el)

		return () => {
			cancelAnimationFrame(frame)
			observer.unobserve(el)
		}
	}, [])

	let body: ReactNode = null
	if (hits.length === 0 && debounced && !isFetching) {
		body = <CommandEmpty>No results found for "{debounced}".</CommandEmpty>
	}

	return (
		<Command
			shouldFilter={false}
			className='search-container bg-background border-border relative h-[495px] w-full overflow-hidden border p-0 backdrop-blur-lg dark:bg-[#0C0C0C]/99'
		>
			<div className='border-border relative border-b'>
				<CommandInput
					autoFocus
					placeholder='Type a command or search...'
					value={term}
					onValueChange={setTerm}
					className='h-[55px] px-4 py-0'
				/>

				{isFetching && (
					<div className='absolute bottom-0 h-[2px] w-full overflow-hidden'>
						<div className='animate-slide-effect absolute top-[1px] h-full w-40 bg-gradient-to-r from-gray-200 via-black via-80% to-gray-200 dark:from-gray-800 dark:via-white dark:via-80% dark:to-gray-800' />
					</div>
				)}
			</div>

			<div className='global-search-list px-2' ref={wrapper}>
				<CommandList ref={list} className='scrollbar-hide'>
					{body}

					{matchingShortcuts.length > 0 && (
						<CommandGroup heading={groupLabels.shortcut}>
							{matchingShortcuts.map(shortcut => (
								<CommandItem
									key={shortcut.id}
									value={shortcut.id}
									onSelect={shortcut.action}
									className='group/item flex flex-col items-start gap-1 py-2 text-sm'
								>
									<div className='flex w-full items-center gap-2'>
										<SearchIcon className='text-dim size-4 shrink-0' />
										{shortcut.title}
									</div>
								</CommandItem>
							))}
						</CommandGroup>
					)}

					{Object.entries(grouped).map(([kind, items]) => (
						<CommandGroup key={kind} heading={groupLabels[kind]}>
							{items.map((hit, index) => (
								<CommandItem
									key={`${hit.kind}-${hit.id}`}
									value={`${hit.kind}-${hit.id}`}
									onSelect={() => {
										openHit(hit)
									}}
									className='group/item flex cursor-pointer flex-col items-start gap-1 py-2 text-sm'
								>
									<HitRow hit={hit} index={index} />
								</CommandItem>
							))}
						</CommandGroup>
					))}
				</CommandList>
			</div>
		</Command>
	)
}
