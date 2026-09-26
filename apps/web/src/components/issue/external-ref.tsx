import { Badge } from '@thom/ui/badge'
import { Link2 } from 'lucide-react'

export const ExternalRef = ({ value }: { readonly value: string }) => (
	<Badge title={`External ref: ${value}`} className='font-mono'>
		<Link2 aria-hidden size={11} />
		{value}
	</Badge>
)
