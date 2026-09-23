import { create } from '_/lib/store'
import { ISSUE_KIND } from '_/core/domain/issue'

type Preview = {
	readonly kind: typeof ISSUE_KIND.TODO
	readonly project: string
	readonly id: string
}

type PreviewState = {
	preview: Preview | undefined
	openPreview: (preview: Preview) => void
	closePreview: () => void
}

export const usePreviewStore = create<PreviewState>()(set => ({
	preview: undefined,
	openPreview: preview => set({ preview }),
	closePreview: () => set({ preview: undefined }),
}))
