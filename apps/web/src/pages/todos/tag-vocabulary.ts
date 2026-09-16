// Mirrors the CLI's list in apps/cli/cmd/folio/tags.go. The API stores
// whatever it is handed, so the vocabulary is the client's opinion: offering a
// fixed list here keeps a typo from becoming a tag nothing will ever query.
export const CONTEXT_TAGS = [
	'api',
	'backend',
	'cli',
	'db',
	'design',
	'docs',
	'frontend',
	'infra',
	'mcp',
	'mobile',
	'tui',
	'web',
] as const

export const KIND_TAGS = [
	'bug',
	'chore',
	'decision',
	'deploy',
	'dx',
	'perf',
	'refactor',
	'release',
	'security',
	'spike',
	'test',
] as const
