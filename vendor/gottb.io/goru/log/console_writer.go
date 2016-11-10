package log

import (
	"fmt"
	"sync"
)

type ConsoleWriter struct {
	mutex *sync.Mutex
}

func NewConsoleWriter() Writer {
	return &ConsoleWriter{&sync.Mutex{}}
}

func (w *ConsoleWriter) Write(msg string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	fmt.Println(msg)
}

func (w *ConsoleWriter) Close() {
}
