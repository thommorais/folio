import { Check, Copy, Link2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Button } from '@thom/ui/button'
import { Input } from '@thom/ui/input'
import { Sheet, SheetContent, SheetHeader, SheetTrigger } from '@thom/ui/sheet'
import { toast } from '@thom/ui/toast'
import { useShareLinks } from '_/app/use-share-links'
import { SHARE_LABEL_MAX, shareUrl, type ShareLink, type ShareTarget } from '_/core/domain/share'
import { Status } from '_/lib/async-status'

const formatDate = (date: Date): string =>
	date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })

const formatVisit = (date: Date): string =>
	date.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })

const copyLink = async (token: string): Promise<boolean> => {
	try {
		await navigator.clipboard.writeText(shareUrl(window.location.origin, token))
		return true
	} catch {
		return false
	}
}

const CopyButton = ({ token }: { readonly token: string }) => {
	const [copied, setCopied] = useState(false)

	useEffect(() => {
		if (!copied) return

		const timer = setTimeout(() => setCopied(false), 2000)

		return () => clearTimeout(timer)
	}, [copied])

	return (
		<Button
			variant='ghost'
			size='sm'
			onClick={async () => {
				if (await copyLink(token)) setCopied(true)
				else toast.error('Could not copy the link.')
			}}
		>
			{copied ? <Check data-slot='icon' /> : <Copy data-slot='icon' />}
			{copied ? 'Copied' : 'Copy'}
		</Button>
	)
}

type RowProps = {
	readonly link: ShareLink
	readonly onRevoke: () => Promise<void>
}

const LinkRow = ({ link, onRevoke }: RowProps) => {
	const [confirming, setConfirming] = useState(false)
	const [revoking, setRevoking] = useState(false)

	return (
		<li className='flex items-center gap-3 px-4 py-3'>
			<div className='min-w-0 flex-1 space-y-0.5'>
				<p className='truncate text-sm'>{link.label}</p>
				<p className='text-dimmer text-xs'>
					Created {formatDate(link.createdAt)} ·{' '}
					{link.lastAccessedAt ? `opened ${formatVisit(link.lastAccessedAt)}` : 'never opened'}
				</p>
			</div>

			{confirming ? (
				<>
					<Button variant='ghost' size='sm' onClick={() => setConfirming(false)}>
						Keep
					</Button>
					<Button
						variant='destructive'
						size='sm'
						loading={revoking}
						onClick={async () => {
							setRevoking(true)
							await onRevoke()
							setRevoking(false)
							setConfirming(false)
						}}
					>
						Revoke
					</Button>
				</>
			) : (
				<>
					<CopyButton token={link.token} />
					<Button variant='ghost' size='sm' onClick={() => setConfirming(true)}>
						Revoke
					</Button>
				</>
			)}
		</li>
	)
}

const SharePanel = ({ target }: { readonly target: ShareTarget }) => {
	const { state, create, revoke } = useShareLinks(target)
	const [label, setLabel] = useState('')
	const [error, setError] = useState<string>()
	const [creating, setCreating] = useState(false)

	const what = target.kind === 'issue' ? 'ticket, with its todos, plans, docs and journal,' : 'plan and its todos'

	const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()
		if (label.trim() === '') return

		setCreating(true)
		setError(undefined)
		const result = await create(label)
		setCreating(false)

		if (!result.success) {
			setError(result.error.message)
			return
		}

		setLabel('')
		if (await copyLink(result.value.token)) toast.success('Link created and copied.')
		else toast.success('Link created.')
	}

	return (
		<div className='flex h-full flex-col gap-6'>
			<SheetHeader>
				<h2 className='font-serif text-lg'>Share</h2>
				<p className='text-dim text-sm'>
					Anyone with a link can read this {what} without signing in, until you revoke it.
				</p>
			</SheetHeader>

			<form onSubmit={onSubmit} className='flex gap-2'>
				<Input
					type='text'
					name='label'
					value={label}
					onChange={event => setLabel(event.target.value)}
					placeholder='Who is this link for?'
					maxLength={SHARE_LABEL_MAX}
					required
					aria-invalid={error ? true : undefined}
				/>
				<Button type='submit' loading={creating} disabled={label.trim() === ''} className='shrink-0'>
					Create link
				</Button>
			</form>

			{error && <p className='text-destructive text-sm'>{error}</p>}

			<section className='space-y-2'>
				<h3 className='text-dim text-xs tracking-wide uppercase'>Links</h3>
				{state.status === Status.Loading && <div className='bg-accent/40 h-16 animate-pulse' />}
				{state.status === Status.Failed && <p className='text-destructive text-sm'>{state.message}</p>}
				{state.status === Status.Ready && state.data.length === 0 && <p className='text-dim text-sm'>No links yet.</p>}
				{state.status === Status.Ready && state.data.length > 0 && (
					<ul className='border-border divide-border divide-y border'>
						{state.data.map(link => (
							<LinkRow
								key={link.id}
								link={link}
								onRevoke={async () => {
									const result = await revoke(link.id)
									if (result.success) toast.success(`Revoked the link for ${link.label}.`)
									else toast.error(result.error.message)
								}}
							/>
						))}
					</ul>
				)}
			</section>
		</div>
	)
}

export const ShareSheet = ({ target }: { readonly target: ShareTarget }) => {
	const [open, setOpen] = useState(false)

	return (
		<Sheet open={open} onOpenChange={setOpen}>
			<SheetTrigger asChild>
				<Button variant='outline' size='sm'>
					<Link2 data-slot='icon' />
					Share
				</Button>
			</SheetTrigger>
			<SheetContent title='Share'>{open && <SharePanel target={target} />}</SheetContent>
		</Sheet>
	)
}
