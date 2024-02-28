package task

type Task interface {
	Namespace() string
	Id() string
	BaseUrl() string
}
