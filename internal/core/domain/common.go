package domain

type Identity struct {
	ID          string
	Name        string
	Description string
}

type IdentityVersion struct {
	Identity
	Version string
}
