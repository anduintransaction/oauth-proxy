package watcher

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gottb.io/goru/errors"
	"gottb.io/goru/utils"

	fsnotify "gopkg.in/fsnotify.v1"
)

type Event struct {
	Name string
	Op   Op
}

func (e Event) String() string {
	// Use a buffer for efficient string concatenation
	var buffer bytes.Buffer

	if e.Op&Create == Create {
		buffer.WriteString("|CREATE")
	}
	if e.Op&Remove == Remove {
		buffer.WriteString("|REMOVE")
	}
	if e.Op&Write == Write {
		buffer.WriteString("|WRITE")
	}

	// If buffer remains empty, return no event names
	if buffer.Len() == 0 {
		return fmt.Sprintf("%q: ", e.Name)
	}

	// Return a list of event names, with leading pipe character stripped
	return fmt.Sprintf("%q: %s", e.Name, buffer.String()[1:])
}

type Op fsnotify.Op

const (
	Create Op = 1 << iota
	Write
	Remove
	rename
	chmod
)

type Watcher struct {
	Events   chan Event
	Errors   chan error
	watchers map[string]*fsnotify.Watcher
	srcTree  map[string]*fileInfo
	mutex    *sync.Mutex
}

func NewWatcher() *Watcher {
	return &Watcher{
		Events:   make(chan Event),
		Errors:   make(chan error),
		watchers: make(map[string]*fsnotify.Watcher),
		srcTree:  make(map[string]*fileInfo),
		mutex:    &sync.Mutex{},
	}
}

func (w *Watcher) add(name string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	f, err := os.Open(name)
	if err != nil {
		return errors.Wrap(err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return errors.Wrap(err)
	}
	var hash string
	ignoreRules := utils.IgnoreRules{}
	if !stat.IsDir() {
		hash, err = hashFile(f)
		if err != nil {
			return err
		}
	} else {
		hash = ""
		ignoreRules = utils.LoadIgnoreRule(name, ".watcher_ignore")
	}
	if stat.IsDir() {
		_, ok := w.watchers[name]
		if !ok {
			fsWatcher, err := fsnotify.NewWatcher()
			if err != nil {
				return errors.Wrap(err)
			}
			err = fsWatcher.Add(name)
			if err != nil {
				return errors.Wrap(err)
			}
			w.watchers[name] = fsWatcher
			go w.watch(name, fsWatcher)
		}
	}
	w.srcTree[name] = &fileInfo{
		stat:        stat,
		hash:        hash,
		ignoreRules: ignoreRules,
	}
	return nil
}

func (w *Watcher) watch(name string, fsWatcher *fsnotify.Watcher) {
	quit := false
	for !quit {
		select {
		case e := <-fsWatcher.Events:
			w.handleEvent(name, e)
		case err := <-fsWatcher.Errors:
			if err != nil {
				w.Errors <- err
			}
			quit = true
		}
	}
}

func (w *Watcher) handleEvent(watcherName string, e fsnotify.Event) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if e.Name == "" || e.Name == watcherName || e.Op == fsnotify.Chmod {
		return
	}
	basename := filepath.Base(e.Name)
	if string(basename[0]) == "." {
		return
	}
	parent := w.srcTree[filepath.Dir(e.Name)]
	if parent != nil && parent.ignoreRules.IsIgnored(basename, ".watcher_ignore") {
		return
	}
	info := w.srcTree[e.Name]
	if e.Op&fsnotify.Write == fsnotify.Write && info != nil && !info.stat.IsDir() {
		f, err := os.Open(e.Name)
		if err != nil {
			w.Errors <- err
			return
		}
		hash, err := hashFile(f)
		if err != nil {
			w.Errors <- err
			return
		}
		if hash == info.hash {
			return
		} else {
			info.hash = hash
		}
	}
	if e.Op&fsnotify.Rename == fsnotify.Rename || e.Op&fsnotify.Remove == fsnotify.Remove {
		e.Op = fsnotify.Remove
		for name := range w.srcTree {
			if strings.HasPrefix(name, e.Name) {
				delete(w.srcTree, name)
			}
		}
		for name, fsWatcher := range w.watchers {
			if strings.HasPrefix(name, e.Name) {
				fsWatcher.Close()
				delete(w.watchers, name)
			}
		}
	}
	if e.Op == fsnotify.Create {
		w.mutex.Unlock()
		err := w.Add(e.Name)
		if err != nil {
			w.Errors <- err
			return
		}
		w.mutex.Lock()
	}
	w.Events <- Event{
		Name: e.Name,
		Op:   Op(e.Op),
	}
}

func (w *Watcher) Add(name string) error {
	return utils.Walk(name, func(path string, info os.FileInfo) error {
		if info == nil {
			return nil
		}
		basename := filepath.Base(path)
		if strings.HasPrefix(basename, ".") {
			return nil
		}
		return w.add(path)
	}, ".watcher_ignore")
}

func (w *Watcher) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	done := make(chan bool, len(w.watchers))
	for _, fsWatcher := range w.watchers {
		go func(fsw *fsnotify.Watcher) {
			fsw.Close()
			done <- true
		}(fsWatcher)
	}
	for i := 0; i < len(w.watchers); i++ {
		<-done
	}
	close(w.Events)
	close(w.Errors)
	return nil
}

func hashFile(f io.Reader) (string, error) {
	h := md5.New()
	_, err := io.Copy(h, f)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type fileInfo struct {
	stat        os.FileInfo
	hash        string
	ignoreRules utils.IgnoreRules
}
