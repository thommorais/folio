import { rounds, currentRound, type Question, type Sent, type StagedMap } from '_/core/domain/interview'
import { markFor } from './marks'

type Props = {
	readonly questions: readonly Question[]
	readonly selected: string | undefined
	readonly staged: StagedMap
	readonly sent: Sent | undefined
	readonly onSelect: (id: string) => void
}

export const Nav = ({ questions, selected, staged, sent, onSelect }: Props) => {
	const latest = currentRound(questions)

	return (
		<nav className='border-border min-h-0 overflow-y-auto border-r pr-2' aria-label='Questions'>
			{questions.length === 0 && <p className='text-dim px-3 py-2 text-sm'>No questions yet.</p>}

			{rounds(questions).map(([round, group]) => (
				<section key={round} className='mb-4'>
					<h2 className='text-dim px-3 py-1 text-xs'>
						Round {round}
						{round === latest ? ' (current)' : ''}
					</h2>

					<ul>
						{group.map(question => {
							const mark = markFor(question, staged[question.id], sent)

							return (
								<li key={question.id}>
									<button
										type='button'
										onClick={() => onSelect(question.id)}
										aria-current={question.id === selected ? 'true' : undefined}
										className='hover:bg-accent/40 aria-[current=true]:bg-accent/60 flex w-full flex-col gap-0.5 px-3 py-2 text-left transition-colors'
									>
										<span className='text-sm'>{question.title}</span>
										<span className='text-dim flex items-center gap-2 text-xs'>
											<span>{question.id}</span>
											<span className={mark.tone === 'live' ? 'text-foreground' : undefined}>{mark.label}</span>
										</span>
									</button>
								</li>
							)
						})}
					</ul>
				</section>
			))}
		</nav>
	)
}
