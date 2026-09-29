import { Button } from '@thom/ui/button'
import {
	isExploring,
	pendingActions,
	stageThread,
	unstageThreadAt,
	type Question,
	type Sent,
	type Staged,
	type StagedMap,
} from '_/core/domain/interview'
import { Field } from './field'

type Props = {
	readonly question: Question
	readonly staged: Staged | undefined
	readonly sent: Sent | undefined
	readonly draft: string
	readonly locked: boolean
	readonly onDraft: (value: string) => void
	readonly onStage: (change: (map: StagedMap) => StagedMap) => void
}

const time = (at: Date): string => at.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

const ExploreTable = ({ question }: { readonly question: Question }) => {
	if (!question.explore) return null

	return (
		<section className='space-y-2'>
			<h3 className='text-sm font-medium'>
				Pros and cons <span className='text-dim font-normal'>{time(question.explore.at)}</span>
			</h3>
			<div className='border-border divide-border divide-y border'>
				<div className='text-dim grid grid-cols-[2rem_1fr_1fr] gap-3 px-3 py-2 text-xs'>
					<span />
					<span>Pros</span>
					<span>Cons</span>
				</div>
				{question.explore.rows.map(row => (
					<div key={row.option} className='grid grid-cols-[2rem_1fr_1fr] gap-3 px-3 py-2 text-xs'>
						<span className={row.option === question.rec.option ? 'text-foreground' : 'text-dim'}>{row.option}</span>
						<ul className='list-disc space-y-1 pl-3'>
							{row.pros.map(pro => (
								<li key={pro}>{pro}</li>
							))}
						</ul>
						<ul className='text-dim list-disc space-y-1 pl-3'>
							{row.cons.map(con => (
								<li key={con}>{con}</li>
							))}
						</ul>
					</div>
				))}
			</div>
		</section>
	)
}

export const Discussion = ({ question, staged, sent, draft, locked, onDraft, onStage }: Props) => {
	const sending = pendingActions(sent, question.id).filter(action => action.type === 'thread')
	const exploring = isExploring(sent, question.id)
	const empty = question.thread.length === 0 && sending.length === 0 && !staged?.thread.length

	const stage = () => {
		const text = draft.trim()

		if (!text) return

		onStage(map => stageThread(map, question.id, text))
		onDraft('')
	}

	return (
		<aside className='border-border flex flex-col gap-6 border-l pl-6' aria-label='Discussion'>
			<ExploreTable question={question} />

			{exploring && (
				<p className='text-dim text-sm'>Exploring: the agent is writing a pros and cons table for this question.</p>
			)}

			<section className='space-y-3'>
				<h3 className='text-sm font-medium'>
					Thread <span className='text-dim font-normal'>({question.thread.length})</span>
				</h3>

				{empty && (
					<p className='text-dim text-sm'>Nothing yet. Ask anything about this question; the agent answers here.</p>
				)}

				<ul className='space-y-3'>
					{question.thread.map((message, index) => (
						<li key={`${message.at.getTime()}-${index}`} className='space-y-0.5'>
							<p className='text-dim text-xs'>
								{message.who} {time(message.at)}
							</p>
							<p className='text-sm whitespace-pre-wrap'>{message.text}</p>
						</li>
					))}

					{sending.map((action, index) => (
						<li key={`sending-${index}`} className='space-y-0.5 opacity-70'>
							<p className='text-dim text-xs'>you, sending</p>
							<p className='text-sm whitespace-pre-wrap'>{action.type === 'thread' ? action.text : ''}</p>
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

			{!locked && (
				<div className='space-y-2'>
					<Field value={draft} disabled={false} placeholder={`Dig deeper on ${question.id}`} onChange={onDraft} />
					<Button size='sm' variant='outline' disabled={!draft.trim()} onClick={stage}>
						Add to discussion
					</Button>
				</div>
			)}
		</aside>
	)
}
