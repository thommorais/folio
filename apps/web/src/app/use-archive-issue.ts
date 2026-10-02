import { useCallback } from 'react'
import type { Result } from '_/lib/result'
import { useContainer } from './container'

export const useArchiveIssue = () => {
	const { issues } = useContainer()

	return useCallback((id: string, archived: boolean): Promise<Result<void>> => issues.setArchived(id, archived), [issues])
}
