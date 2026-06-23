package models

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func (r Role) CanInvite() bool {
	return r == RoleOwner || r == RoleAdmin
}

func (r Role) CanManageTask() bool {
	return r == RoleOwner || r == RoleAdmin
}

func (r Role) CanDeleteTeam() bool {
	return r == RoleOwner
}

func (r Role) Valid() bool {
	return r == RoleOwner || r == RoleAdmin || r == RoleMember
}
