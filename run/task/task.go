package task

type Task interface {
	Id() int
	Namespace() string
	BaseUrl() string
}
