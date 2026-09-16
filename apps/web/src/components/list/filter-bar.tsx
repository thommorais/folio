import { cn } from '@thom/libs/cn'
import {
	DropdownMenu,
	DropdownMenuCheckboxItem,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuPortal,
	DropdownMenuSub,
	DropdownMenuSubContent,
	DropdownMenuSubTrigger,
	DropdownMenuTrigger,
} from '@thom/ui/dropdown-menu'
import { useState } from 'react'
import { ClearIcon, FilterIcon, SearchIcon } from './icons'

const FilterMenuItem = ({ label, children }: { readonly label: string; readonly children: React.ReactNode }) => (
	<DropdownMenuGroup>
		<DropdownMenuSub>
			<DropdownMenuSubTrigger>
				<span>{label}</span>
			</DropdownMenuSubTrigger>
			<DropdownMenuPortal>
				<DropdownMenuSubContent sideOffset={14} alignOffset={-4} className='p-0'>
					{children}
				</DropdownMenuSubContent>
			</DropdownMenuPortal>
		</DropdownMenuSub>
	</DropdownMenuGroup>
)

const FilterCheckboxItem = ({
	label,
	checked,
	onCheckedChange,
}: {
	readonly label: string
	readonly checked: boolean
	readonly onCheckedChange: () => void
}) => (
	<DropdownMenuCheckboxItem
		checked={checked}
		onCheckedChange={onCheckedChange}
		// Selecting closes the menu by default, which would end a multi-select
		// after the first choice.
		onSelect={event => {
			event.preventDefault()
		}}
	>
		{label}
	</DropdownMenuCheckboxItem>
)

const Chip = ({ label, onRemove }: { readonly label: string; readonly onRemove: () => void }) => (
	<button
		type='button'
		onClick={onRemove}
		className='bg-secondary text-dim group flex h-9 max-w-full items-center gap-1 px-2 text-sm font-normal'
	>
		<ClearIcon className='w-0 shrink-0 scale-0 transition-all group-hover:w-4 group-hover:scale-100' />
		<span className='truncate'>{label}</span>
	</button>
)

const toggle = <T,>(values: readonly T[] | undefined, value: T): readonly T[] | undefined => {
	const current = values ?? []
	const next = current.includes(value) ? current.filter(entry => entry !== value) : [...current, value]

	return next.length > 0 ? next : undefined
}

type ActiveFilter = {
	readonly key: string
	readonly label: string
	readonly onRemove: () => void
}

type Props = {
	readonly placeholder: string
	readonly term: string | undefined
	readonly onSearch: (term: string | undefined) => void
	readonly chips: readonly ActiveFilter[]
	readonly children: React.ReactNode
	readonly trailing?: React.ReactNode
}

const FilterBar = ({ placeholder, term, onSearch, chips, children, trailing }: Props) => {
	const [draft, setDraft] = useState(term ?? '')

	return (
		<DropdownMenu>
			<div className='flex flex-wrap items-center gap-2'>
				<form
					className='relative w-full sm:w-auto'
					onSubmit={event => {
						event.preventDefault()
						onSearch(draft === '' ? undefined : draft)
					}}
				>
					<span className='text-dim pointer-events-none absolute top-2.5 left-3'>
						<SearchIcon />
					</span>

					<input
						value={draft}
						onChange={event => {
							setDraft(event.target.value)
							if (event.target.value === '') {
								onSearch(undefined)
							}
						}}
						placeholder={placeholder}
						autoComplete='off'
						autoCapitalize='none'
						autoCorrect='off'
						spellCheck={false}
						className='border-border h-9 w-full border bg-transparent pr-9 pl-9 text-sm focus:outline-hidden sm:w-[320px]'
					/>

					<DropdownMenuTrigger asChild>
						<button
							type='button'
							aria-label='Filters'
							className={cn(
								'absolute top-2.5 right-3 z-10 opacity-50 transition-opacity duration-300 hover:opacity-100',
								chips.length > 0 && 'opacity-100',
							)}
						>
							<FilterIcon />
						</button>
					</DropdownMenuTrigger>
				</form>

				{chips.map(chip => (
					<Chip key={chip.key} label={chip.label} onRemove={chip.onRemove} />
				))}

				{trailing !== undefined && <div className='ml-auto'>{trailing}</div>}
			</div>

			<DropdownMenuContent className='w-[220px]' align='end' sideOffset={19} alignOffset={-11} side='bottom'>
				{children}
			</DropdownMenuContent>
		</DropdownMenu>
	)
}

export { FilterBar, FilterCheckboxItem, FilterMenuItem, toggle, type ActiveFilter }
