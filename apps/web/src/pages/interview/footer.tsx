import { useState } from 'react'
import { Button } from '@thom/ui/button'
import { summarize, type StagedMap } from '_/core/domain/interview'

type Props = {
	readonly staged: StagedMap
	readonly count: number
	readonly canSend: boolean
	readonly canFinish: boolean
	readonly sending: boolean
	readonly pending: boolean
	readonly working: boolean
	readonly error: string | undefined
	readonly onSend: (finish: boolean) => void
}

const status = ({ count, staged, pending, working, error }: Props): string => {
	if (error) return error
	if (pending) return 'Sent. Tell the agent to continue.'
	if (working) return 'The agent is working. Send unlocks when it is waiting.'

	return count === 0 ? 'Nothing staged. Pick an option or write in the discussion.' : `Staged: ${summarize(staged)}.`
}

export const Footer = (props: Props) => {
	const { canSend, canFinish, sending, count, onSend, error } = props
	const [confirming, setConfirming] = useState(false)

	return (
		<footer className='border-border bg-background sticky bottom-0 flex items-center justify-between gap-4 border-t py-4'>
			<p className={error ? 'text-destructive text-sm' : 'text-dim text-sm'} role='status'>
				{status(props)}
			</p>

			<div className='flex items-center gap-2'>
				{confirming ? (
					<>
						<span className='text-dim text-sm'>Close the interview after this Send?</span>
						<Button
							size='sm'
							loading={sending}
							disabled={!canSend}
							onClick={() => {
								setConfirming(false)
								onSend(true)
							}}
						>
							Finish
						</Button>
						<Button size='sm' variant='ghost' onClick={() => setConfirming(false)}>
							Cancel
						</Button>
					</>
				) : (
					<>
						<Button
							size='sm'
							variant='outline'
							disabled={!canFinish}
							title={canFinish ? undefined : 'Answer or defer every open question first'}
							onClick={() => setConfirming(true)}
						>
							Finish
						</Button>
						<Button size='sm' loading={sending} disabled={!canSend || count === 0} onClick={() => onSend(false)}>
							{count > 0 ? `Send ${count}` : 'Send'}
						</Button>
					</>
				)}
			</div>
		</footer>
	)
}
