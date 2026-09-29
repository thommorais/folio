import { isOpen, type Question, type Sent, type Staged } from '_/core/domain/interview'

export type Mark = { readonly label: string; readonly tone: 'plain' | 'live' }

export const answerLabel = (question: Question): string => {
	const { answer, rec } = question

	if (!answer) return ''
	if (answer.kind === 'accept') return rec.option ?? 'accepted'
	if (answer.kind === 'option') return answer.option ?? ''

	return 'text'
}

const sentTouches = (sent: Sent | undefined, id: string): boolean =>
	sent?.actions.some(action => 'q' in action && action.q === id) ?? false

export const markFor = (question: Question, staged: Staged | undefined, sent: Sent | undefined): Mark => {
	if (staged) return { label: 'staged', tone: 'live' }
	if (sentTouches(sent, question.id)) return { label: 'sent', tone: 'live' }
	if (question.status === 'answered') return { label: answerLabel(question), tone: 'plain' }
	if (question.status === 'deferred') return { label: 'deferred', tone: 'plain' }
	if (question.status === 'reopened') return { label: 'reopened', tone: 'live' }

	return { label: isOpen(question) ? 'open' : question.status, tone: 'plain' }
}
