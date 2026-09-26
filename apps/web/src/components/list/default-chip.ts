type Chip = {
	readonly key: string
	readonly label: string
	readonly onRemove: () => void
}

export const defaultChip = <S extends string>(
	key: string,
	chosen: readonly S[] | undefined,
	all: readonly S[],
	shown: readonly S[],
	label: (status: S) => string,
	showAll: (all: readonly S[]) => void,
): Chip | undefined => {
	if (chosen !== undefined) return undefined

	const hidden = all.filter(status => !shown.includes(status))
	if (hidden.length === 0) return undefined

	return {
		key: `${key}-default`,
		label: `Hiding ${hidden.map(label).join(', ')}`,
		onRemove: () => showAll(all),
	}
}
