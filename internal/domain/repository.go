package domain

type ServiceRepository interface {
	JobRepository
	TaskRunRepository
}
