package domain

import (
	"time"
)

type Job struct {
	Identity
	Status        string
	Progress      float32
	SubmittedDate time.Time
	StartDate     time.Time
	EndDate       time.Time
}
