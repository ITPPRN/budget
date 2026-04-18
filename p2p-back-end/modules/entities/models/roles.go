package models

const (
	RoleAdmin          = "ADMIN"           // Global Admin
	RoleOwner          = "OWNER"           // Department Owner (Manage Dept)
	RoleDelegate       = "DELEGATE"        // Department Delegate (Assistant)
	RoleBranchDelegate = "BRANCH_DELEGATE" // Like DELEGATE but data scoped to user's branch
)
