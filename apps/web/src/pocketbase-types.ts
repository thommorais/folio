/**
* This file was @generated using pocketbase-typegen
*/

import type PocketBase from 'pocketbase'
import type { RecordService } from 'pocketbase'

export const Collections = {
	Authorigins: "_authOrigins",
	Externalauths: "_externalAuths",
	Mfas: "_mfas",
	Otps: "_otps",
	Superusers: "_superusers",
	JournClients: "journ_clients",
	JournCycles: "journ_cycles",
	JournDomains: "journ_domains",
	JournEntries: "journ_entries",
	JournEntryTags: "journ_entry_tags",
	JournIssueLinks: "journ_issue_links",
	JournIssueTags: "journ_issue_tags",
	JournIssues: "journ_issues",
	JournMembers: "journ_members",
	JournPlans: "journ_plans",
	JournProjects: "journ_projects",
	JournTags: "journ_tags",
	Users: "users",
} as const
export type Collections = typeof Collections[keyof typeof Collections]

// Alias types for improved usability
export type IsoDateString = string
export type IsoAutoDateString = string & { readonly autodate: unique symbol }
export type RecordIdString = string
export type FileNameString = string & { readonly filename: unique symbol }
export type HTMLString = string

type ExpandType<T> = unknown extends T
	? T extends unknown
		? { expand?: unknown }
		: { expand: T }
	: { expand: T }

// System fields
export type BaseSystemFields<T = unknown> = {
	id: RecordIdString
	collectionId: string
	collectionName: Collections
} & ExpandType<T>

export type AuthSystemFields<T = unknown> = {
	email: string
	emailVisibility: boolean
	username: string
	verified: boolean
} & BaseSystemFields<T>

// Record types for each collection

export type AuthoriginsRecord = {
	collectionRef: string
	created: IsoAutoDateString
	fingerprint: string
	id: string
	recordRef: string
	updated: IsoAutoDateString
}

export type ExternalauthsRecord = {
	collectionRef: string
	created: IsoAutoDateString
	id: string
	provider: string
	providerId: string
	recordRef: string
	updated: IsoAutoDateString
}

export type MfasRecord = {
	collectionRef: string
	created: IsoAutoDateString
	id: string
	method: string
	recordRef: string
	updated: IsoAutoDateString
}

export type OtpsRecord = {
	collectionRef: string
	created: IsoAutoDateString
	id: string
	password: string
	recordRef: string
	sentTo?: string
	updated: IsoAutoDateString
}

export type SuperusersRecord = {
	created: IsoAutoDateString
	email: string
	emailVisibility?: boolean
	id: string
	password: string
	tokenKey: string
	updated: IsoAutoDateString
	verified?: boolean
}

export type JournClientsRecord = {
	created: IsoAutoDateString
	descr?: string
	id: string
	logo?: string
	name: string
	site?: string
	slug: string
	updated: IsoAutoDateString
}

export const JournCyclesPhaseOptions = {
	"plan": "plan",
	"do": "do",
	"check": "check",
	"act": "act",
} as const
export type JournCyclesPhaseOptions = typeof JournCyclesPhaseOptions[keyof typeof JournCyclesPhaseOptions]
export type JournCyclesRecord = {
	closed_at?: IsoDateString
	created: IsoAutoDateString
	created_by?: RecordIdString
	id: string
	issue: RecordIdString
	ordinal: number
	phase: JournCyclesPhaseOptions
	project: RecordIdString
	resolution?: string
	updated: IsoAutoDateString
}

export type JournDomainsRecord = {
	client: RecordIdString
	created: IsoAutoDateString
	descr?: string
	id: string
	name: string
	slug: string
	updated: IsoAutoDateString
}

export const JournEntriesKindOptions = {
	"journal": "journal",
	"doc": "doc",
	"log": "log",
} as const
export type JournEntriesKindOptions = typeof JournEntriesKindOptions[keyof typeof JournEntriesKindOptions]
export type JournEntriesRecord<Tmeta = unknown, Ttags = unknown> = {
	body?: HTMLString
	branch?: string
	created: IsoAutoDateString
	created_by?: RecordIdString
	cycle?: RecordIdString
	domain: RecordIdString
	external_ref?: string
	id: string
	issue?: RecordIdString
	kind: JournEntriesKindOptions
	meta?: null | Tmeta
	plan?: RecordIdString
	pr?: string
	project: RecordIdString
	slug?: string
	tags?: null | Ttags
	title?: string
	updated: IsoAutoDateString
}

export type JournEntryTagsRecord = {
	created: IsoAutoDateString
	entry: RecordIdString
	id: string
	tag: RecordIdString
	updated: IsoAutoDateString
}

export const JournIssueLinksKindOptions = {
	"blocks": "blocks",
	"relates": "relates",
	"parent": "parent",
} as const
export type JournIssueLinksKindOptions = typeof JournIssueLinksKindOptions[keyof typeof JournIssueLinksKindOptions]
export type JournIssueLinksRecord = {
	created: IsoAutoDateString
	domain: RecordIdString
	from: RecordIdString
	id: string
	kind: JournIssueLinksKindOptions
	to: RecordIdString
	updated: IsoAutoDateString
}

export type JournIssueTagsRecord = {
	created: IsoAutoDateString
	id: string
	issue: RecordIdString
	tag: RecordIdString
	updated: IsoAutoDateString
}

export const JournIssuesKindOptions = {
	"ticket": "ticket",
	"todo": "todo",
} as const
export type JournIssuesKindOptions = typeof JournIssuesKindOptions[keyof typeof JournIssuesKindOptions]

export const JournIssuesStatusOptions = {
	"open": "open",
	"in_progress": "in_progress",
	"blocked": "blocked",
	"done": "done",
	"cancelled": "cancelled",
} as const
export type JournIssuesStatusOptions = typeof JournIssuesStatusOptions[keyof typeof JournIssuesStatusOptions]

export const JournIssuesPriorityOptions = {
	"low": "low",
	"medium": "medium",
	"high": "high",
} as const
export type JournIssuesPriorityOptions = typeof JournIssuesPriorityOptions[keyof typeof JournIssuesPriorityOptions]

export const JournIssuesWayfinderOptions = {
	"map": "map",
	"research": "research",
	"prototype": "prototype",
	"grilling": "grilling",
	"task": "task",
} as const
export type JournIssuesWayfinderOptions = typeof JournIssuesWayfinderOptions[keyof typeof JournIssuesWayfinderOptions]
export type JournIssuesRecord<Ttags = unknown> = {
	assignee?: RecordIdString
	body?: HTMLString
	created: IsoAutoDateString
	created_by?: RecordIdString
	domain: RecordIdString
	due_date?: IsoDateString
	external_ref?: string
	id: string
	kind: JournIssuesKindOptions
	plan?: RecordIdString
	position?: number
	priority: JournIssuesPriorityOptions
	project: RecordIdString
	size?: number
	slug: string
	status: JournIssuesStatusOptions
	tags?: null | Ttags
	title: string
	updated: IsoAutoDateString
	wayfinder?: JournIssuesWayfinderOptions
}

export const JournMembersRoleOptions = {
	"owner": "owner",
	"editor": "editor",
	"viewer": "viewer",
} as const
export type JournMembersRoleOptions = typeof JournMembersRoleOptions[keyof typeof JournMembersRoleOptions]
export type JournMembersRecord = {
	created: IsoAutoDateString
	domain: RecordIdString
	id: string
	project: RecordIdString
	role: JournMembersRoleOptions
	updated: IsoAutoDateString
	user: RecordIdString
}

export const JournPlansStatusOptions = {
	"draft": "draft",
	"active": "active",
	"done": "done",
	"abandoned": "abandoned",
} as const
export type JournPlansStatusOptions = typeof JournPlansStatusOptions[keyof typeof JournPlansStatusOptions]
export type JournPlansRecord<Ttags = unknown> = {
	created: IsoAutoDateString
	created_by?: RecordIdString
	goal?: string
	id: string
	issue?: RecordIdString
	project: RecordIdString
	status: JournPlansStatusOptions
	tags?: null | Ttags
	title: string
	updated: IsoAutoDateString
}

export type JournProjectsRecord = {
	archived?: boolean
	created: IsoAutoDateString
	descr?: string
	domain?: RecordIdString
	id: string
	name: string
	slug: string
	updated: IsoAutoDateString
}

export type JournTagsRecord = {
	created: IsoAutoDateString
	domain: RecordIdString
	id: string
	name: string
	slug: string
	updated: IsoAutoDateString
}

export type UsersRecord = {
	avatar?: FileNameString
	created: IsoAutoDateString
	email: string
	emailVisibility?: boolean
	id: string
	name?: string
	password: string
	tokenKey: string
	updated: IsoAutoDateString
	verified?: boolean
}

// Response types include system fields and match responses from the PocketBase API
export type AuthoriginsResponse<Texpand = unknown> = Required<AuthoriginsRecord> & BaseSystemFields<Texpand>
export type ExternalauthsResponse<Texpand = unknown> = Required<ExternalauthsRecord> & BaseSystemFields<Texpand>
export type MfasResponse<Texpand = unknown> = Required<MfasRecord> & BaseSystemFields<Texpand>
export type OtpsResponse<Texpand = unknown> = Required<OtpsRecord> & BaseSystemFields<Texpand>
export type SuperusersResponse<Texpand = unknown> = Required<SuperusersRecord> & AuthSystemFields<Texpand>
export type JournClientsResponse<Texpand = unknown> = Required<JournClientsRecord> & BaseSystemFields<Texpand>
export type JournCyclesResponse<Texpand = unknown> = Required<JournCyclesRecord> & BaseSystemFields<Texpand>
export type JournDomainsResponse<Texpand = unknown> = Required<JournDomainsRecord> & BaseSystemFields<Texpand>
export type JournEntriesResponse<Tmeta = unknown, Ttags = unknown, Texpand = unknown> = Required<JournEntriesRecord<Tmeta, Ttags>> & BaseSystemFields<Texpand>
export type JournEntryTagsResponse<Texpand = unknown> = Required<JournEntryTagsRecord> & BaseSystemFields<Texpand>
export type JournIssueLinksResponse<Texpand = unknown> = Required<JournIssueLinksRecord> & BaseSystemFields<Texpand>
export type JournIssueTagsResponse<Texpand = unknown> = Required<JournIssueTagsRecord> & BaseSystemFields<Texpand>
export type JournIssuesResponse<Ttags = unknown, Texpand = unknown> = Required<JournIssuesRecord<Ttags>> & BaseSystemFields<Texpand>
export type JournMembersResponse<Texpand = unknown> = Required<JournMembersRecord> & BaseSystemFields<Texpand>
export type JournPlansResponse<Ttags = unknown, Texpand = unknown> = Required<JournPlansRecord<Ttags>> & BaseSystemFields<Texpand>
export type JournProjectsResponse<Texpand = unknown> = Required<JournProjectsRecord> & BaseSystemFields<Texpand>
export type JournTagsResponse<Texpand = unknown> = Required<JournTagsRecord> & BaseSystemFields<Texpand>
export type UsersResponse<Texpand = unknown> = Required<UsersRecord> & AuthSystemFields<Texpand>

// Types containing all Records and Responses, useful for creating typing helper functions

export type CollectionRecords = {
	_authOrigins: AuthoriginsRecord
	_externalAuths: ExternalauthsRecord
	_mfas: MfasRecord
	_otps: OtpsRecord
	_superusers: SuperusersRecord
	journ_clients: JournClientsRecord
	journ_cycles: JournCyclesRecord
	journ_domains: JournDomainsRecord
	journ_entries: JournEntriesRecord
	journ_entry_tags: JournEntryTagsRecord
	journ_issue_links: JournIssueLinksRecord
	journ_issue_tags: JournIssueTagsRecord
	journ_issues: JournIssuesRecord
	journ_members: JournMembersRecord
	journ_plans: JournPlansRecord
	journ_projects: JournProjectsRecord
	journ_tags: JournTagsRecord
	users: UsersRecord
}

export type CollectionResponses = {
	_authOrigins: AuthoriginsResponse
	_externalAuths: ExternalauthsResponse
	_mfas: MfasResponse
	_otps: OtpsResponse
	_superusers: SuperusersResponse
	journ_clients: JournClientsResponse
	journ_cycles: JournCyclesResponse
	journ_domains: JournDomainsResponse
	journ_entries: JournEntriesResponse
	journ_entry_tags: JournEntryTagsResponse
	journ_issue_links: JournIssueLinksResponse
	journ_issue_tags: JournIssueTagsResponse
	journ_issues: JournIssuesResponse
	journ_members: JournMembersResponse
	journ_plans: JournPlansResponse
	journ_projects: JournProjectsResponse
	journ_tags: JournTagsResponse
	users: UsersResponse
}

// Utility types for create/update operations

type ProcessCreateAndUpdateFields<T> = Omit<{
	// Omit AutoDate fields
	[K in keyof T as Extract<T[K], IsoAutoDateString> extends never ? K : never]: 
		// Convert FileNameString to File
		T[K] extends infer U ? 
			U extends (FileNameString | FileNameString[]) ? 
				U extends any[] ? File[] : File 
			: U
		: never
}, 'id'>

// Create type for Auth collections
export type CreateAuth<T> = {
	id?: RecordIdString
	email: string
	emailVisibility?: boolean
	password: string
	passwordConfirm: string
	verified?: boolean
} & ProcessCreateAndUpdateFields<T>

// Create type for Base collections
export type CreateBase<T> = {
	id?: RecordIdString
} & ProcessCreateAndUpdateFields<T>

// Update type for Auth collections
export type UpdateAuth<T> = Partial<
	Omit<ProcessCreateAndUpdateFields<T>, keyof AuthSystemFields>
> & {
	email?: string
	emailVisibility?: boolean
	oldPassword?: string
	password?: string
	passwordConfirm?: string
	verified?: boolean
}

// Update type for Base collections
export type UpdateBase<T> = Partial<
	Omit<ProcessCreateAndUpdateFields<T>, keyof BaseSystemFields>
>

// Get the correct create type for any collection
export type Create<T extends keyof CollectionResponses> =
	CollectionResponses[T] extends AuthSystemFields
		? CreateAuth<CollectionRecords[T]>
		: CreateBase<CollectionRecords[T]>

// Get the correct update type for any collection
export type Update<T extends keyof CollectionResponses> =
	CollectionResponses[T] extends AuthSystemFields
		? UpdateAuth<CollectionRecords[T]>
		: UpdateBase<CollectionRecords[T]>

// Type for usage with type asserted PocketBase instance
// https://github.com/pocketbase/js-sdk#specify-typescript-definitions

export type TypedPocketBase = {
	collection<T extends keyof CollectionResponses>(
		idOrName: T
	): RecordService<CollectionResponses[T]>
} & PocketBase
