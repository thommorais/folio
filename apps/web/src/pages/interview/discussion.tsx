import { Button } from '@thom/ui/button'
import { stageThread, unstageThreadAt, type Question, type Staged, type StagedMap } from '_/core/domain/interview'

import { Field } from './field'

type Props = {
	readonly question: Question
	readonly staged: Staged | undefined
	readonly draft: string
	readonly locked: boolean
	readonly onDraft: (value: string) => void
	readonly onStage: (change: (map: StagedMap) => StagedMap) => void
}

const ExploreTable = ({ question }: { readonly question: Question }) => {
	if (!question.explore) return null

	return (
		<section className='space-y-2'>
			<h3 className='text-sm font-medium'>Explored</h3>
			<div className='border-border divide-border divide-y border'>
				{question.explore.rows.map(row => (
					<div key={row.option} className='grid grid-cols-[2rem_1fr_1fr] gap-3 px-3 py-2 text-xs'>
						<span className='text-dim'>{row.option}</span>
						<ul className='space-y-1'>
							{row.pros.map(pro => (
								<li key={pro}>+ {pro}</li>
							))}
						</ul>
						<ul className='text-dim space-y-1'>
							{row.cons.map(con => (
								<li key={con}>- {con}</li>
							))}
						</ul>
					</div>
				))}
			</div>
		</section>
	)
}

export const Discussion = ({ question, staged, draft, locked, onDraft, onStage }: Props) => {
	const stage = () => {
		const text = draft.trim()

		if (!text) return

		onStage(map => stageThread(map, question.id, text))
		onDraft('')
	}

	return (
		<aside className='border-border flex min-h-0 flex-col gap-6 overflow-y-auto border-l pl-6' aria-label='Discussion'>
			<ExploreTable question={question} />

			<section className='space-y-3'>
				<h3 className='text-sm font-medium'>
					Thread <span className='text-dim font-normal'>({question.thread.length})</span>
				</h3>

				{question.thread.length === 0 && !staged?.thread.length && (
					<p className='text-dim text-sm'>No messages on this question.</p>
				)}

				<ul className='space-y-3'>
					{question.thread.map((message, index) => (
						<li key={`${message.at.getTime()}-${index}`} className='space-y-0.5'>
							<p className='text-dim text-xs'>{message.who}</p>
							<p className='text-sm whitespace-pre-wrap'>{message.text}</p>
						</li>
					))}

					{staged?.thread.map((text, index) => (
						<li key={`staged-${index}`} className='border-border space-y-0.5 border border-dashed px-3 py-2'>
							<p className='text-dim flex items-center justify-between text-xs'>
								<span>you, staged</span>
								<button
									type='button'
									className='hover:text-foreground underline'
									onClick={() => onStage(map => unstageThreadAt(map, question.id, index))}
								>
									Remove
								</button>
							</p>
							<p className='text-sm whitespace-pre-wrap'>{text}</p>
						</li>
					))}
				</ul>
			</section>

			<div className='space-y-2'>
				<Field value={draft} disabled={locked} placeholder='Message the agent about this question' onChange={onDraft} />
				<Button size='sm' variant='outline' disabled={locked || !draft.trim()} onClick={stage}>
					Stage message
				</Button>
			</div>
		</aside>
	)
}
