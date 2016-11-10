package log

type Writer interface {
	Write(msg string)
	Close()
}
