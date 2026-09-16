import { create } from '_/lib/store'

type SearchState = {
	isOpen: boolean
	setOpen: (open?: boolean) => void
}

export const useSearchStore = create<SearchState>()(set => ({
	isOpen: false,
	setOpen: open => set(state => ({ isOpen: open ?? !state.isOpen })),
}))
