import type { IssueStatus } from '_/core/domain/issue'

export const ISSUE_STATUS_LABELS: Record<IssueStatus, string> = {
	open: 'Open',
	in_progress: 'In progress',
	blocked: 'Blocked',
	done: 'Done',
	cancelled: 'Cancelled',
}
