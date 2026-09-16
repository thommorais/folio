import { create } from '_/lib/store'

type Preview = {
	readonly kind: 'todo'
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
