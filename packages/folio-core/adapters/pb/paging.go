package pb

// applyPaging trims an already-loaded slice. PocketBase's FindAllRecords has
// no limit argument, so paging happens here; the service clamps the limit so
// the slice stays bounded.
func applyPaging[T any](items []T, offset, limit int) []T {
	if offset > 0 {
		if offset >= len(items) {
			return items[:0]
		}
		items = items[offset:]
	}
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items
}
