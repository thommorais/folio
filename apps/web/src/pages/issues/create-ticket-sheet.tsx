import { Plus } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@thom/ui/button'
import { Input } from '@thom/ui/input'
import { Sheet, SheetContent, SheetHeader, SheetTrigger } from '@thom/ui/sheet'
import { toast } from '@thom/ui/toast'
import { useContainer } from '_/app/container'
import { ISSUE_KIND, PRIORITIES, PRIORITY, SIZES, isSize, type Priority } from '_/core/domain/issue'

const controlClasses =
	'border-border text-foreground placeholder:text-muted-foreground focus-visible:border-ring w-full border bg-transparent px-3 text-sm transition-colors focus-visible:outline-none'

const FieldLabel = ({ htmlFor, children }: { readonly htmlFor: string; readonly children: React.ReactNode }) => (
	<label htmlFor={htmlFor} className='block text-sm font-medium'>
		{children}
	</label>
)

const Hint = ({ children }: { readonly children: React.ReactNode }) => <p className='text-dim text-xs'>{children}</p>

const CreateTicketForm = ({ project, onCreated }: { readonly project: string; readonly onCreated: () => void }) => {
	const { issues } = useContainer()
	const [title, setTitle] = useState('')
	const [body, setBody] = useState('')
	const [priority, setPriority] = useState<Priority>(PRIORITY.MEDIUM)
	const [size, setSize] = useState('')
	const [error, setError] = useState<string>()
	const [creating, setCreating] = useState(false)

	const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()
		if (title.trim() === '') return

		const points = Number(size)

		setCreating(true)
		setError(undefined)
		const result = await issues.create(project, {
			kind: ISSUE_KIND.TICKET,
			title: title.trim(),
			body: body.trim() === '' ? undefined : body,
			priority,
			size: isSize(points) ? points : undefined,
		})
		setCreating(false)

		if (!result.success) {
			setError(result.error.message)
			return
		}

		toast.success(`Created ${result.value.slug}.`)
		onCreated()
	}

	return (
		<form onSubmit={onSubmit} className='flex h-full flex-col gap-8'>
			<SheetHeader>
				<h2 className='font-serif text-lg'>Create ticket</h2>
			</SheetHeader>

			<div className='min-h-0 flex-1 space-y-8 overflow-y-auto'>
				<div className='space-y-2'>
					<FieldLabel htmlFor='ticket-title'>Title</FieldLabel>
					<Input
						id='ticket-title'
						type='text'
						value={title}
						onChange={event => setTitle(event.target.value)}
						placeholder='What needs doing?'
						autoComplete='off'
						autoFocus
						required
						aria-invalid={error ? true : undefined}
					/>
					<Hint>The slug is derived from it.</Hint>
				</div>

				<div className='space-y-2'>
					<FieldLabel htmlFor='ticket-body'>Description</FieldLabel>
					<textarea
						id='ticket-body'
						value={body}
						onChange={event => setBody(event.target.value)}
						rows={8}
						className={`${controlClasses} block resize-none py-2`}
					/>
					<Hint>Markdown.</Hint>
				</div>

				<div className='flex gap-4'>
					<div className='flex-1 space-y-2'>
						<FieldLabel htmlFor='ticket-priority'>Priority</FieldLabel>
						<select
							id='ticket-priority'
							value={priority}
							onChange={event => setPriority(event.target.value as Priority)}
							className={`${controlClasses} h-9`}
						>
							{PRIORITIES.map(value => (
								<option key={value} value={value}>
									{value}
								</option>
							))}
						</select>
					</div>

					<div className='flex-1 space-y-2'>
						<FieldLabel htmlFor='ticket-size'>Size</FieldLabel>
						<select
							id='ticket-size'
							value={size}
							onChange={event => setSize(event.target.value)}
							className={`${controlClasses} h-9`}
						>
							<option value=''>Unsized</option>
							{SIZES.map(value => (
								<option key={value} value={value}>
									{value}
								</option>
							))}
						</select>
					</div>
				</div>

				{error && <p className='text-destructive text-sm'>{error}</p>}
			</div>

			<Button type='submit' fullWidth loading={creating} disabled={title.trim() === ''}>
				Create
			</Button>
		</form>
	)
}

export const CreateTicketSheet = ({ project }: { readonly project: string }) => {
	const [open, setOpen] = useState(false)

	return (
		<Sheet open={open} onOpenChange={setOpen}>
			<SheetTrigger asChild>
				<Button variant='outline' size='icon' aria-label='Create ticket'>
					<Plus data-slot='icon' />
				</Button>
			</SheetTrigger>
			<SheetContent title='Create ticket'>
				{open && <CreateTicketForm project={project} onCreated={() => setOpen(false)} />}
			</SheetContent>
		</Sheet>
	)
}
