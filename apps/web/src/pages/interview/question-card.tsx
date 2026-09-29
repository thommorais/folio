import { Button } from '@thom/ui/button'
import { Markdown } from '_/components/markdown'
import {
	isExploring,
	isOpen,
	pendingActions,
	stageAnswer,
	stageDefer,
	stageExplore,
	stageReopen,
	unstageDecision,
	unstageKind,
	type Question,
	type Sent,
	type Staged,
	type StagedAnswer,
	type StagedMap,
} from '_/core/domain/interview'
import { Field } from './field'
import { answerText } from './marks'

type Props = {
	readonly question: Question
	readonly staged: Staged | undefined
	readonly sent: Sent | undefined
	readonly draft: string
	readonly locked: boolean
	readonly onDraft: (value: string) => void
	readonly onStage: (change: (map: StagedMap) => StagedMap) => void
	readonly onSelect: (id: string) => void
}

const optionOf = (question: Question, answer: StagedAnswer | undefined): string | undefined => {
	if (!answer) return undefined

	return answer.kind === 'accept' ? question.rec.option : answer.option
}

const stagedLine = (question: Question, staged: Staged | undefined): string => {
	if (staged?.defer) return 'Staged: defer this question'
	if (staged?.reopen) return 'Staged: reopen this question'

	const answer = staged?.answer

	if (!answer) return ''
	if (answer.kind === 'accept') return `Staged: accept ${question.rec.option ?? ''}`
	if (answer.kind === 'option') return `Staged: option ${answer.option ?? ''}`

	return `Staged text: "${answer.text ?? ''}"`
}

export const QuestionCard = ({ question, staged, sent, draft, locked, onDraft, onStage, onSelect }: Props) => {
	const open = isOpen(question)
	const hasOptions = question.options.length > 0
	const pending = pendingActions(sent, question.id)
	const pendingAnswer = pending.findLast(action => action.type === 'answer')
	const pendingDecisions = pending.filter(action => action.type === 'defer' || action.type === 'reopen')
	const exploring = isExploring(sent, question.id)
	const stagedAnswer = staged?.answer
	const chosenKey = stagedAnswer ? optionOf(question, stagedAnswer) : optionOf(question, question.answer)
	const pendingKey =
		!stagedAnswer && pendingAnswer?.type === 'answer'
			? optionOf(question, { kind: pendingAnswer.kind, option: pendingAnswer.option })
			: undefined
	const picked = question.answer !== undefined || stagedAnswer !== undefined || pendingAnswer !== undefined
	const line = stagedLine(question, staged)

	const pick = (key: string) => {
		if (stagedAnswer && optionOf(question, stagedAnswer) === key) {
			onStage(map => unstageKind(map, question.id, 'answer'))
			return
		}

		onStage(map =>
			stageAnswer(
				map,
				question.id,
				key === question.rec.option ? { kind: 'accept', option: key } : { kind: 'option', option: key },
			),
		)
	}

	const stageText = () => {
		const text = draft.trim()

		if (!text) return

		onStage(map => stageAnswer(map, question.id, { kind: 'text', text }))
		onDraft('')
	}

	return (
		<article className='space-y-6 lg:px-6'>
			<header className='space-y-2'>
				<p className='text-dim flex flex-wrap items-center gap-x-3 text-xs'>
					<span>{question.id}</span>
					<span>round {question.round}</span>
					{question.deps.length === 0 ? (
						<span>root</span>
					) : (
						<span>
							after{' '}
							{question.deps.map((dep, index) => (
								<span key={dep}>
									{index > 0 && ', '}
									<button type='button' onClick={() => onSelect(dep)} className='hover:text-foreground underline'>
										{dep}
									</button>
								</span>
							))}
						</span>
					)}
					{question.durable && <span>durable</span>}
					{question.status === 'deferred' && <span>deferred</span>}
					{question.status === 'reopened' && <span>reopened</span>}
					{question.updated && <span className='text-foreground'>recommendation updated</span>}
				</p>

				<h2 className='text-lg font-medium'>{question.title}</h2>
				{question.body && <Markdown>{question.body}</Markdown>}
			</header>

			{hasOptions ? (
				<ul
					className={`border-border divide-border divide-y border transition-opacity ${picked ? '[&>li:not([data-picked])]:opacity-50 [&>li:not([data-picked])]:hover:opacity-100' : ''}`}
					aria-label='Options'
				>
					{question.options.map(option => {
						const recommended = option.k === question.rec.option
						const selected = chosenKey === option.k
						const sending = pendingKey === option.k

						return (
							<li key={option.k} data-picked={selected || sending ? '' : undefined}>
								<button
									type='button'
									disabled={locked}
									onClick={() => pick(option.k)}
									aria-pressed={selected}
									data-sending={sending ? '' : undefined}
									className='hover:bg-accent/40 aria-pressed:bg-accent/60 data-[sending]:border-foreground/40 flex w-full items-start gap-3 px-4 py-3 text-left text-sm transition-colors disabled:opacity-60 data-[sending]:border-l-2 data-[sending]:border-dashed'
								>
									<span className='text-dim w-4 shrink-0 text-xs'>{option.k}</span>
									<span className='flex-1'>{option.text}</span>
									{sending && <span className='text-dim shrink-0 text-xs'>sending</span>}
									{recommended && <span className='text-dim shrink-0 text-xs'>recommended</span>}
								</button>
							</li>
						)
					})}
				</ul>
			) : (
				question.rec.text && (
					<div className='border-border flex items-start justify-between gap-4 border px-4 py-3 text-sm'>
						<span>
							<span className='text-dim text-xs'>Recommended</span>
							<br />
							{question.rec.text}
						</span>
						<Button
							size='sm'
							variant='outline'
							disabled={locked}
							onClick={() =>
								onStage(map => stageAnswer(map, question.id, { kind: 'text', text: question.rec.text ?? '' }))
							}
						>
							Accept
						</Button>
					</div>
				)
			)}

			{question.answer?.kind === 'text' && (
				<p className='border-border border px-4 py-3 text-sm'>
					<span className='text-dim text-xs'>Your answer</span>
					<br />
					{answerText(question)}
				</p>
			)}

			{pendingAnswer?.type === 'answer' && pendingAnswer.kind === 'text' && !stagedAnswer && (
				<p className='border-border border border-dashed px-4 py-3 text-sm'>
					<span className='text-dim text-xs'>Your answer, sending</span>
					<br />
					{pendingAnswer.text}
				</p>
			)}

			{question.rec.why && (
				<p className={`text-dim text-sm leading-relaxed ${picked ? 'opacity-50 hover:opacity-100' : ''}`}>
					<span className='text-xs'>Why</span>
					<br />
					{question.rec.why}
				</p>
			)}

			{line && (
				<p className='text-sm'>
					{line}
					<button
						type='button'
						className='text-dim hover:text-foreground ml-3 text-xs underline'
						onClick={() => onStage(map => unstageDecision(map, question.id))}
					>
						clear
					</button>
				</p>
			)}

			{pendingDecisions.length > 0 && (
				<p className='text-dim text-sm'>
					Sent: {pendingDecisions.map(action => action.type).join(', ')} this question. Waiting for the agent.
				</p>
			)}

			{!locked && (
				<div className='space-y-2'>
					<Field
						value={draft}
						disabled={false}
						placeholder={
							question.status === 'answered'
								? 'Change your answer, or add nuance'
								: hasOptions
									? 'Free-text answer, if none of the options fit'
									: 'Your answer'
						}
						onChange={onDraft}
					/>
					<Button size='sm' variant='outline' disabled={!draft.trim()} onClick={stageText}>
						Stage answer
					</Button>
				</div>
			)}

			<div className='flex flex-wrap items-center gap-2'>
				{open ? (
					<Button
						size='sm'
						variant='outline'
						disabled={locked}
						aria-pressed={staged?.defer === true}
						onClick={() =>
							onStage(map => (staged?.defer ? unstageKind(map, question.id, 'defer') : stageDefer(map, question.id)))
						}
					>
						{staged?.defer ? 'Deferral staged' : 'Defer'}
					</Button>
				) : (
					<Button
						size='sm'
						variant='outline'
						disabled={locked}
						aria-pressed={staged?.reopen === true}
						onClick={() =>
							onStage(map => (staged?.reopen ? unstageKind(map, question.id, 'reopen') : stageReopen(map, question.id)))
						}
					>
						{staged?.reopen ? 'Reopening staged' : 'Reopen'}
					</Button>
				)}

				{hasOptions && (
					<Button
						size='sm'
						variant='ghost'
						disabled={locked || exploring}
						aria-pressed={staged?.explore === true}
						onClick={() =>
							onStage(map =>
								staged?.explore ? unstageKind(map, question.id, 'explore') : stageExplore(map, question.id),
							)
						}
					>
						{exploring
							? 'Exploring'
							: staged?.explore
								? 'Explore staged'
								: question.explore
									? 'Explore again'
									: 'Explore deeper'}
					</Button>
				)}
			</div>
		</article>
	)
}
