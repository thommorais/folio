import { useCallback } from 'react'
import type { Result } from '_/lib/result'
import { useContainer } from './container'

export const useDeleteIssue = () => {
	const { issues } = useContainer()

	return useCallback((id: string): Promise<Result<void>> => issues.remove(id), [issues])
}
