package uds

type Logger interface {
	Error(args ...interface{})
	Debug(args ...interface{})
}