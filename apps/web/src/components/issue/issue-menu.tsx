import { Archive, ArchiveRestore, Link2, MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@thom/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@thom/ui/dropdown-menu'
import { toast } from '@thom/ui/toast'
import { useArchiveIssue } from '_/app/use-archive-issue'
import { useDeleteIssue } from '_/app/use-delete-issue'
import { EditIssueSheet } from '_/components/issue/edit-issue-sheet'
import { ShareSheetPanel } from '_/components/share/share-sheet'
import type { Issue } from '_/core/domain/issue'
import type { ShareTarget } from '_/core/domain/share'

type Props = {
	readonly issue: Issue
	readonly onDeleted: () => void
	readonly onArchived: () => void
	readonly share?: ShareTarget
}

export const IssueMenu = ({ issue, onDeleted, onArchived, share }: Props) => {
	const deleteIssue = useDeleteIssue()
	const archiveIssue = useArchiveIssue()
	const [open, setOpen] = useState(false)
	const [confirming, setConfirming] = useState(false)
	const [deleting, setDeleting] = useState(false)
	const [sharing, setSharing] = useState(false)
	const [editing, setEditing] = useState(false)

	const onOpenChange = (next: boolean) => {
		if (deleting) return
		setOpen(next)
		if (!next) setConfirming(false)
	}

	const onArchive = async () => {
		const archived = !issue.archived
		const result = await archiveIssue(issue.id, archived)

		if (!result.success) {
			toast.error(result.error.message)
			return
		}

		toast.success(archived ? 'Archived.' : 'Restored.')
		if (archived) onArchived()
	}

	const onDelete = async () => {
		setDeleting(true)
		const result = await deleteIssue(issue.id)
		setDeleting(false)

		if (!result.success) {
			toast.error(result.error.message)
			return
		}

		setOpen(false)
		toast.success('Deleted.')
		onDeleted()
	}

	return (
		<>
		<DropdownMenu open={open} onOpenChange={onOpenChange}>
			<DropdownMenuTrigger asChild>
				<Button variant='ghost' size='icon' aria-label='More actions'>
					<MoreHorizontal data-slot='icon' />
				</Button>
			</DropdownMenuTrigger>

			<DropdownMenuContent align='end'>
				{confirming ? (
					<div className='space-y-2 p-2'>
						<p className='text-sm'>Delete this issue?</p>
						<div className='flex justify-end gap-2'>
							<Button variant='ghost' size='sm' disabled={deleting} onClick={() => setConfirming(false)}>
								Keep
							</Button>
							<Button variant='destructive' size='sm' loading={deleting} onClick={onDelete}>
								Delete
							</Button>
						</div>
					</div>
				) : (
					<>
						<DropdownMenuItem className='gap-2' onSelect={() => setEditing(true)}>
							<Pencil className='size-4' />
							Edit
						</DropdownMenuItem>
						{share && (
							<DropdownMenuItem className='gap-2' onSelect={() => setSharing(true)}>
								<Link2 className='size-4' />
								Share
							</DropdownMenuItem>
						)}
						<DropdownMenuItem className='gap-2' onSelect={() => void onArchive()}>
							{issue.archived ? <ArchiveRestore className='size-4' /> : <Archive className='size-4' />}
							{issue.archived ? 'Restore' : 'Archive'}
						</DropdownMenuItem>
						<DropdownMenuItem
							className='text-destructive gap-2'
							onSelect={event => {
								event.preventDefault()
								setConfirming(true)
							}}
						>
							<Trash2 className='size-4' />
							Delete
						</DropdownMenuItem>
					</>
				)}
			</DropdownMenuContent>
		</DropdownMenu>
		<EditIssueSheet issue={issue} open={editing} onOpenChange={setEditing} />
		{share && <ShareSheetPanel target={share} open={sharing} onOpenChange={setSharing} />}
		</>
	)
}
