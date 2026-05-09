package database

type logger interface {
	Info(args ...interface{})
	Warnf(format string, args ...interface{})
	Fatal(args ...interface{})
}
