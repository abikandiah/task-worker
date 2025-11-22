package domain

import (
	"github.com/google/uuid"
)

type Identity struct {
	IdentitySubmission
	ID uuid.UUID
}

type IdentitySubmission struct {
	Name        string
	Description string
}

type IdentityVersion struct {
	Identity
	Version string
}
