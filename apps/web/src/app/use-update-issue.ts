import { useCallback } from 'react'
import type { UpdateIssueInput } from '_/core/ports/issues'
import type { Result } from '_/lib/result'
import { useContainer } from './container'

export const useUpdateIssue = () => {
	const { issues } = useContainer()

	return useCallback((id: string, input: UpdateIssueInput): Promise<Result<void>> => issues.update(id, input), [issues])
}
