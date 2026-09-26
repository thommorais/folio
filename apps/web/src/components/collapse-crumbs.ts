export const collapseCrumbs = <T extends { readonly title: string }>(crumbs: readonly T[]): readonly T[] =>
	crumbs.filter((crumb, index) => crumb.title.toLowerCase() !== crumbs[index + 1]?.title.toLowerCase())
