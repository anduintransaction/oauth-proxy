package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gottb.io/goru/errors"
)

type RollingPeriod int

const (
	RollMinute RollingPeriod = iota
	RollHour
	RollDay
)

type RollingWriter struct {
	mutex        *sync.Mutex
	w            io.Closer
	buf          *bufio.Writer
	filename     string
	period       RollingPeriod
	bytesWritten int
	ticker       *time.Ticker
	done         chan bool
	lastTime     time.Time
}

func NewRollingWriter(filename string, period RollingPeriod) (Writer, error) {
	w := &RollingWriter{
		mutex:    &sync.Mutex{},
		filename: filename,
		period:   period,
		done:     make(chan bool, 1),
	}
	err := w.rollLogFile()
	if err != nil {
		return nil, err
	}
	go w.background()
	w.ticker = time.NewTicker(time.Second)
	return w, nil
}

func (w *RollingWriter) Write(message string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	n, err := w.buf.WriteString(message + "\n")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	w.bytesWritten += n
}

func (w *RollingWriter) Close() {
	if w.w != nil {
		w.w.Close()
	}
	w.ticker.Stop()
	w.done <- true
}

func (w *RollingWriter) rollLogFile() error {
	stat, err := os.Stat(w.filename)
	if stat == nil || err != nil {
		return w.newLogFile()
	}
	now := time.Now()
	modTime := stat.ModTime()
	w.lastTime = modTime
	if w.canBeRolled(modTime, now) {
		if stat.Size() > 0 {
			err = w.renameOldFile(modTime)
			if err != nil {
				return err
			}
		}
		return w.newLogFile()
	} else {
		f, err := os.OpenFile(w.filename, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return errors.Wrap(err)
		}
		w.w = f
		w.buf = bufio.NewWriter(f)
	}
	return nil
}

func (w *RollingWriter) newLogFile() error {
	dir := filepath.Dir(w.filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	f, err := os.Create(w.filename)
	if err != nil {
		return errors.Wrap(err)
	}
	f.Chmod(0644)
	w.w = f
	w.buf = bufio.NewWriter(f)
	w.lastTime = time.Now()
	return nil
}

func (w *RollingWriter) background() {
	for {
		select {
		case now := <-w.ticker.C:
			w.mutex.Lock()
			w.flush()
			w.roll(w.lastTime, now)
			w.lastTime = now
			w.mutex.Unlock()
		case <-w.done:
			return
		}
	}
}

func (w *RollingWriter) flush() {
	w.buf.Flush()
}

func (w *RollingWriter) roll(lastTime, now time.Time) {
	if !w.canBeRolled(lastTime, now) {
		return
	}
	w.w.Close()
	if w.bytesWritten > 0 {
		err := w.renameOldFile(lastTime)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	err := w.newLogFile()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	w.bytesWritten = 0
}

func (w *RollingWriter) canBeRolled(lastTime, now time.Time) bool {
	ly, lM, ld := lastTime.Date()
	lh, lm, _ := lastTime.Clock()
	y, M, d := now.Date()
	h, m, _ := now.Clock()
	if ly != y || lM != M || ld != d {
		return true
	}
	switch w.period {
	case RollMinute:
		return lh != h || lm != m
	case RollHour:
		return lh != h
	}
	return false
}

func (w *RollingWriter) renameOldFile(lastTime time.Time) error {
	var suffix string
	switch w.period {
	case RollMinute:
		suffix = lastTime.Format("2006-01-02-15-04")
	case RollHour:
		suffix = lastTime.Format("2006-01-02-15")
	default:
		suffix = lastTime.Format("2006-01-02")
	}
	newFilename := w.filename + "." + suffix
	return errors.Wrap(os.Rename(w.filename, newFilename))
}
