import { useState } from 'react'
import { Button } from '@thom/ui/button'
import { Input } from '@thom/ui/input'
import { Sheet, SheetContent, SheetHeader } from '@thom/ui/sheet'
import { toast } from '@thom/ui/toast'
import { useUpdateIssue } from '_/app/use-update-issue'
import { controlClasses, FieldLabel, Hint } from '_/components/issue/form-fields'
import { ISSUE_KIND, PRIORITIES, SIZES, isSize, type Issue, type Priority } from '_/core/domain/issue'

type FormProps = {
	readonly issue: Issue
	readonly onSaved: () => void
}

const EditIssueForm = ({ issue, onSaved }: FormProps) => {
	const updateIssue = useUpdateIssue()
	const [title, setTitle] = useState(issue.title)
	const [body, setBody] = useState(issue.body)
	const [priority, setPriority] = useState<Priority>(issue.priority)
	const [size, setSize] = useState(issue.size === undefined ? '' : String(issue.size))
	const [error, setError] = useState<string>()
	const [saving, setSaving] = useState(false)

	const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()
		if (title.trim() === '') return

		const points = Number(size)

		setSaving(true)
		setError(undefined)
		const result = await updateIssue(issue.id, {
			title: title.trim(),
			body,
			priority,
			size: isSize(points) ? points : null,
		})
		setSaving(false)

		if (!result.success) {
			setError(result.error.message)
			return
		}

		toast.success('Saved.')
		onSaved()
	}

	const noun = issue.kind === ISSUE_KIND.TICKET ? 'ticket' : 'todo'

	return (
		<form onSubmit={onSubmit} className='flex h-full flex-col gap-8'>
			<SheetHeader>
				<h2 className='font-serif text-lg'>Edit {noun}</h2>
			</SheetHeader>

			<div className='min-h-0 flex-1 space-y-8 overflow-y-auto'>
				<div className='space-y-2'>
					<FieldLabel htmlFor='issue-title'>Title</FieldLabel>
					<Input
						id='issue-title'
						type='text'
						value={title}
						onChange={event => setTitle(event.target.value)}
						autoComplete='off'
						autoFocus
						required
						aria-invalid={error ? true : undefined}
					/>
					<Hint>The slug stays {issue.slug}.</Hint>
				</div>

				<div className='space-y-2'>
					<FieldLabel htmlFor='issue-body'>Description</FieldLabel>
					<textarea
						id='issue-body'
						value={body}
						onChange={event => setBody(event.target.value)}
						rows={8}
						className={`${controlClasses} block resize-none py-2`}
					/>
					<Hint>Markdown.</Hint>
				</div>

				<div className='flex gap-4'>
					<div className='flex-1 space-y-2'>
						<FieldLabel htmlFor='issue-priority'>Priority</FieldLabel>
						<select
							id='issue-priority'
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
						<FieldLabel htmlFor='issue-size'>Size</FieldLabel>
						<select
							id='issue-size'
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

			<Button type='submit' fullWidth loading={saving} disabled={title.trim() === ''}>
				Save
			</Button>
		</form>
	)
}

type Props = {
	readonly issue: Issue
	readonly open: boolean
	readonly onOpenChange: (open: boolean) => void
}

export const EditIssueSheet = ({ issue, open, onOpenChange }: Props) => (
	<Sheet open={open} onOpenChange={onOpenChange}>
		<SheetContent title='Edit'>{open && <EditIssueForm issue={issue} onSaved={() => onOpenChange(false)} />}</SheetContent>
	</Sheet>
)
