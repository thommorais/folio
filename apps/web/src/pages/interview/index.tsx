import { useState } from 'react'
import { useHotkeys } from 'react-hotkeys-hook'
import { Link, useParams } from '@tanstack/react-router'
import { LoadError } from '_/components/load-error'
import { RecordGone } from '_/components/record/record-gone'
import { useInterview } from '_/app/use-interview'
import { useInterviewSession } from '_/app/use-interview-session'
import { useIssue } from '_/app/use-issue'
import { firstOpenQuestion, isStuck, openCount, type Interview } from '_/core/domain/interview'
import { Status } from '_/lib/async-status'
import { Discussion } from './discussion'
import { Footer } from './footer'
import { Nav } from './nav'
import { QuestionCard } from './question-card'
import { Terms } from './terms'

type Params = { readonly client: string; readonly domain: string; readonly slug: string; readonly ticket: string }

const InterviewPage = () => {
	const { client, domain, slug, ticket } = useParams({
		from: '/_authenticated/$client/$domain/$slug/grilling/$ticket',
	}) satisfies Params

	const issue = useIssue(slug, ticket)

	if (issue.status === Status.Idle || issue.status === Status.Loading) return null
	if (issue.status === Status.Failed) return <LoadError message={issue.message} />

	if (issue.status === Status.Gone) {
		return (
			<RecordGone title={issue.title}>
				<Link to='/$client/$domain/$slug/tickets' params={{ client, domain, slug }} className='text-sm underline'>
					Back to tickets
				</Link>
			</RecordGone>
		)
	}

	return <Loaded client={client} domain={domain} slug={slug} ticketSlug={ticket} issueId={issue.issue.id} />
}

type LoadedProps = {
	readonly client: string
	readonly domain: string
	readonly slug: string
	readonly ticketSlug: string
	readonly issueId: string
}

const Loaded = ({ client, domain, slug, ticketSlug, issueId }: LoadedProps) => {
	const state = useInterview(issueId)

	if (state.status === Status.Idle || state.status === Status.Loading) return null
	if (state.status === Status.Failed) return <LoadError message={state.message} />

	if (state.status === Status.Gone) {
		return (
			<RecordGone title={state.title}>
				<Link
					to='/$client/$domain/$slug/tickets/$ticket'
					params={{ client, domain, slug, ticket: ticketSlug }}
					className='text-sm underline'
				>
					Back to the ticket
				</Link>
			</RecordGone>
		)
	}

	return <Room interview={state.interview} client={client} domain={domain} slug={slug} ticketSlug={ticketSlug} />
}

type RoomProps = {
	readonly interview: Interview
	readonly client: string
	readonly domain: string
	readonly slug: string
	readonly ticketSlug: string
}

const Room = ({ interview, client, domain, slug, ticketSlug }: RoomProps) => {
	const session = useInterviewSession(interview)
	const [chosen, setChosen] = useState<string | undefined>(undefined)
	const [showTerms, setShowTerms] = useState(false)

	const { questions, agentStatus, agentSince, finishedAt } = interview
	const selected = questions.find(question => question.id === chosen) ?? firstOpenQuestion(questions)
	const open = openCount(questions)
	const locked = finishedAt !== undefined
	const now = Date.now()
	const working = agentStatus === 'working' && !isStuck(agentSince, now)
	const waitingOnAgent = session.pending !== undefined && !isStuck(new Date(session.pending.at), now)
	const canSend = !locked && !session.sending && !working && !waitingOnAgent
	const settled = questions.length - open

	useHotkeys('mod+enter', () => void session.send(false), {
		enableOnFormTags: true,
		enabled: canSend && session.count > 0,
	})

	return (
		<div className='flex flex-col gap-4'>
			<header className='border-border flex items-center justify-between gap-4 border-b pb-4'>
				<div className='min-w-0'>
					<Link
						to='/$client/$domain/$slug/tickets/$ticket'
						params={{ client, domain, slug, ticket: ticketSlug }}
						className='text-dim hover:text-foreground text-xs transition-colors'
					>
						Back to the ticket
					</Link>
					<h1 className='truncate text-base font-medium'>{interview.topic}</h1>
				</div>

				<div className='text-dim flex shrink-0 items-center gap-4 text-xs'>
					<button
						type='button'
						onClick={() => setShowTerms(value => !value)}
						aria-pressed={showTerms}
						className='hover:text-foreground aria-pressed:text-foreground transition-colors'
					>
						Terms ({interview.terms.length})
					</button>
					<span>
						{locked
							? 'finished'
							: agentStatus === 'working'
								? `agent working since ${agentSince.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
								: `waiting for you${interview.handled > 0 ? `, handled #${interview.handled}` : ''}`}
					</span>
					<span>
						{settled} of {questions.length} settled
					</span>
				</div>
			</header>

			<div className='grid grid-cols-1 gap-6 lg:grid-cols-[260px_minmax(0,2fr)_minmax(0,1fr)]'>
				<Nav
					questions={questions}
					selected={selected?.id}
					staged={session.staged}
					sent={session.pending}
					onSelect={setChosen}
				/>

				<div className='space-y-6'>
					{locked && (
						<p className='border-border mx-6 border px-4 py-3 text-sm'>
							Finished {finishedAt.toLocaleString()}. The interview is closed.
						</p>
					)}

					{interview.note && <p className='text-dim mx-6 text-sm'>{interview.note}</p>}

					{showTerms && <Terms terms={interview.terms} />}

					{selected ? (
						<QuestionCard
							key={selected.id}
							question={selected}
							staged={session.staged[selected.id]}
							sent={session.pending}
							draft={session.drafts[selected.id]?.text ?? ''}
							locked={locked}
							onDraft={value => session.setDraft(selected.id, 'text', value)}
							onStage={session.setStaged}
							onSelect={setChosen}
						/>
					) : (
						<p className='text-dim px-6 text-sm'>The agent has not posted a round yet.</p>
					)}
				</div>

				{selected && (
					<Discussion
						key={selected.id}
						question={selected}
						staged={session.staged[selected.id]}
						sent={session.pending}
						draft={session.drafts[selected.id]?.thread ?? ''}
						locked={locked}
						onDraft={value => session.setDraft(selected.id, 'thread', value)}
						onStage={session.setStaged}
					/>
				)}
			</div>

			<Footer
				staged={session.staged}
				count={session.count}
				canSend={canSend}
				canFinish={canSend && open === 0 && questions.length > 0}
				sending={session.sending}
				pending={session.pending !== undefined}
				working={working}
				locked={locked}
				error={session.error}
				onSend={finish => void session.send(finish)}
			/>
		</div>
	)
}

export { InterviewPage }
