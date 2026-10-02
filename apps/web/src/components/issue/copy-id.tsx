import { Check, Copy } from 'lucide-react'
import { Button } from '@thom/ui/button'
import { toast } from '@thom/ui/toast'
import { useCopy } from '_/app/use-copy'

export const CopyId = ({ id }: { readonly id: string }) => {
	const { copied, copy } = useCopy()

	return (
		<Button
			variant='outline'
			size='sm'
			onClick={() => {
				copy(id).catch(() => toast.error('Could not copy the ID.'))
			}}
		>
			{copied ? <Check data-slot='icon' /> : <Copy data-slot='icon' />}
			{copied ? 'Copied' : 'Copy ID'}
		</Button>
	)
}
