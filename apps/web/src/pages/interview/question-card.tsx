import { Button } from '@thom/ui/button'
import { Markdown } from '_/components/markdown'
import {
	isOpen,
	stageAnswer,
	stageDefer,
	stageExplore,
	stageReopen,
	unstageKind,
	type Question,
	type Staged,
	type StagedMap,
} from '_/core/domain/interview'
import { Field } from './field'
import { answerText } from './marks'

type Props = {
	readonly question: Question
	readonly staged: Staged | undefined
	readonly draft: string
	readonly locked: boolean
	readonly onDraft: (value: string) => void
	readonly onStage: (change: (map: StagedMap) => StagedMap) => void
	readonly onSelect: (id: string) => void
}

const chosenOption = (question: Question, staged: Staged | undefined): string | undefined => {
	const source = staged?.answer ?? (staged ? undefined : question.answer)

	if (!source) return undefined
	if (source.kind === 'accept') return question.rec.option
	if (source.kind === 'option') return source.option

	return undefined
}

export const QuestionCard = ({ question, staged, draft, locked, onDraft, onStage, onSelect }: Props) => {
	const open = isOpen(question)
	const chosen = chosenOption(question, staged)
	const hasOptions = question.options.length > 0
	const stagedText = staged?.answer?.kind === 'text' ? staged.answer.text : undefined

	const pick = (key: string) => {
		if (chosen === key && staged?.answer) {
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
	}

	return (
		<article className='space-y-6 px-6'>
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
					{question.updated && <span className='text-foreground'>recommendation updated</span>}
				</p>

				<h2 className='text-lg font-medium'>{question.title}</h2>
				{question.body && <Markdown>{question.body}</Markdown>}
			</header>

			{!open && question.answer && (
				<p className='border-border border px-4 py-3 text-sm'>
					<span className='text-dim text-xs'>Answer</span>
					<br />
					{answerText(question)}
				</p>
			)}

			{hasOptions ? (
				<ul className='border-border divide-border divide-y border' aria-label='Options'>
					{question.options.map(option => {
						const recommended = option.k === question.rec.option
						const selected = chosen === option.k

						return (
							<li key={option.k}>
								<button
									type='button'
									disabled={locked || (!open && !staged)}
									onClick={() => pick(option.k)}
									aria-pressed={selected}
									className='hover:bg-accent/40 aria-pressed:bg-accent/60 flex w-full items-start gap-3 px-4 py-3 text-left text-sm transition-colors disabled:opacity-60'
								>
									<span className='text-dim w-4 shrink-0 text-xs'>{option.k}</span>
									<span className='flex-1'>{option.text}</span>
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
							disabled={locked || !open}
							onClick={() => onStage(map => stageAnswer(map, question.id, { kind: 'text', text: question.rec.text }))}
						>
							Accept
						</Button>
					</div>
				)
			)}

			{question.rec.why && (
				<p className='text-dim text-sm leading-relaxed'>
					<span className='text-xs'>Why</span>
					<br />
					{question.rec.why}
				</p>
			)}

			{(open || staged?.answer) && (
				<div className='space-y-2'>
					<Field value={draft} disabled={locked} placeholder='Your own answer' onChange={onDraft} />
					<div className='flex items-center gap-3'>
						<Button size='sm' variant='outline' disabled={locked || !draft.trim()} onClick={stageText}>
							Stage answer
						</Button>
						{stagedText && <span className='text-dim text-xs'>Staged: {stagedText}</span>}
					</div>
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
						disabled={locked}
						aria-pressed={staged?.explore === true}
						onClick={() =>
							onStage(map =>
								staged?.explore ? unstageKind(map, question.id, 'explore') : stageExplore(map, question.id),
							)
						}
					>
						{staged?.explore ? 'Explore staged' : 'Explore deeper'}
					</Button>
				)}
			</div>
		</article>
	)
}
