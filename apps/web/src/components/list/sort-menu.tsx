import {
	DropdownMenu,
	DropdownMenuCheckboxItem,
	DropdownMenuContent,
	DropdownMenuTrigger,
} from '@thom/ui/dropdown-menu'
import type { Sort, SortDirection } from '_/core/ports/sort'
import { SortIcon } from './icons'
import { nextSort } from './next-sort'

const ARROW: Record<SortDirection, string> = { asc: '↑', desc: '↓' }

type Props<TField extends string> = {
	readonly fields: readonly TField[]
	readonly labels: Record<TField, string>
	readonly sort: Sort<TField> | undefined
	readonly onChange: (sort: Sort<TField> | undefined) => void
	readonly defaultLabel: string
}

const SortMenu = <TField extends string>({ fields, labels, sort, onChange, defaultLabel }: Props<TField>) => (
	<DropdownMenu>
		<DropdownMenuTrigger asChild>
			<button
				type='button'
				aria-label='Sort'
				className='border-border text-dim hover:text-foreground flex h-9 items-center gap-2 border px-3 text-sm'
			>
				<SortIcon direction={sort?.direction} />
				<span className='hidden sm:inline'>{sort ? `${labels[sort.field]} ${ARROW[sort.direction]}` : defaultLabel}</span>
			</button>
		</DropdownMenuTrigger>

		<DropdownMenuContent className='w-[180px]' align='end' sideOffset={6}>
			{fields.map(field => (
				<DropdownMenuCheckboxItem
					key={field}
					checked={sort?.field === field}
					onCheckedChange={() => {
						onChange(nextSort(sort, field))
					}}
					onSelect={event => {
						event.preventDefault()
					}}
				>
					<span className='flex w-full items-center justify-between gap-2'>
						<span>{labels[field]}</span>
						{sort?.field === field && <span className='text-dim'>{ARROW[sort.direction]}</span>}
					</span>
				</DropdownMenuCheckboxItem>
			))}

			{sort !== undefined && (
				<DropdownMenuCheckboxItem
					checked={false}
					onCheckedChange={() => {
						onChange(undefined)
					}}
				>
					<span className='text-dim'>Reset to {defaultLabel.toLowerCase()}</span>
				</DropdownMenuCheckboxItem>
			)}
		</DropdownMenuContent>
	</DropdownMenu>
)

export { SortMenu }
