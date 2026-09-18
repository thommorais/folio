import { Sheet, SheetContent } from '@thom/ui/sheet'
import { usePreviewStore } from '_/app/preview-store'
import { IssueDetails } from '_/pages/issues/issue-details'

export const PreviewSheet = () => {
	const preview = usePreviewStore(state => state.preview)
	const closePreview = usePreviewStore(state => state.closePreview)

	return (
		<Sheet
			open={preview !== undefined}
			onOpenChange={next => {
				if (!next) closePreview()
			}}
		>
			<SheetContent title='Todo details'>
				{preview && <IssueDetails project={preview.project} todoId={preview.id} />}
			</SheetContent>
		</Sheet>
	)
}
