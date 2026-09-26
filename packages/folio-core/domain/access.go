package domain

type DomainGrant struct {
	DomainID DomainID
	Role     Role
	Members  []UserID
}

type ProjectAccess struct {
	ClientOwners []UserID
	DomainGrants []DomainGrant
	Members      []Member
}
