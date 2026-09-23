import {
	DropdownMenu,
	DropdownMenuCheckboxItem,
	DropdownMenuContent,
	DropdownMenuTrigger,
} from '@thom/ui/dropdown-menu'
import { SORT_DIRECTION, type Sort, type SortDirection } from '_/core/ports/sort'
import { SortIcon } from './icons'

// Clicking the active field flips direction; clicking another switches to it
// descending, which is what a date or priority column is usually wanted in.
const next = <TField extends string>(current: Sort<TField> | undefined, field: TField): Sort<TField> => {
	if (current?.field !== field) {
		return { field, direction: SORT_DIRECTION.DESC }
	}
	return { field, direction: current.direction === SORT_DIRECTION.DESC ? SORT_DIRECTION.ASC : SORT_DIRECTION.DESC }
}

const ARROW: Record<SortDirection, string> = { asc: '↑', desc: '↓' }

type Props<TField extends string> = {
	readonly fields: readonly TField[]
	readonly labels: Record<TField, string>
	readonly sort: Sort<TField> | undefined
	readonly onChange: (sort: Sort<TField> | undefined) => void
}

const SortMenu = <TField extends string>({ fields, labels, sort, onChange }: Props<TField>) => (
	<DropdownMenu>
		<DropdownMenuTrigger asChild>
			<button
				type='button'
				aria-label='Sort'
				className='border-border text-dim hover:text-foreground flex h-9 items-center gap-2 border px-3 text-sm'
			>
				<SortIcon direction={sort?.direction} />
				<span className='hidden sm:inline'>{sort ? labels[sort.field] : 'Sort'}</span>
			</button>
		</DropdownMenuTrigger>

		<DropdownMenuContent className='w-[180px]' align='end' sideOffset={6}>
			{fields.map(field => (
				<DropdownMenuCheckboxItem
					key={field}
					checked={sort?.field === field}
					onCheckedChange={() => {
						onChange(next(sort, field))
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
					<span className='text-dim'>Clear sort</span>
				</DropdownMenuCheckboxItem>
			)}
		</DropdownMenuContent>
	</DropdownMenu>
)

export { SortMenu }
