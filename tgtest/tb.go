package tgtest

type TB interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Fatal(args ...interface{})
}
