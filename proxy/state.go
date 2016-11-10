package proxy

import (
	"net/http"
	"sync"
	"time"

	"gottb.io/goru/log"
)

var defaultStateMap *stateMap

func AddState(name string, proxy *Proxy, request *http.Request) {
	defaultStateMap.add(name, proxy, request)
}

func GetState(name string) *State {
	return defaultStateMap.get(name)
}

func AcquireState(name string) *State {
	return defaultStateMap.acquire(name)
}

type UserInfo struct {
	Name  string `json:"login"`
	Email string `json:"email"`
}

type Session struct {
	User    *UserInfo `json:"user"`
	Version int64     `json:"version"`
}

type State struct {
	Name    string
	Proxy   *Proxy
	Request *http.Request
	User    *UserInfo
}

type stateMap struct {
	mutex  sync.Mutex
	white  map[string]*State
	gray   map[string]*State
	ticker *time.Ticker
	stop   chan bool
}

func newStateMap(stateTimeout int) *stateMap {
	m := &stateMap{
		white:  make(map[string]*State),
		gray:   make(map[string]*State),
		ticker: time.NewTicker(time.Duration(stateTimeout) * time.Second),
		stop:   make(chan bool, 1),
	}
	go m.gc()
	return m
}

func (m *stateMap) add(name string, proxy *Proxy, request *http.Request) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	log.Debugf("State add: %s to %s", name, request.URL.String())
	m.white[name] = &State{
		Name:    name,
		Proxy:   proxy,
		Request: request,
	}
}

func (m *stateMap) get(name string) *State {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.getUnsafe(name)
}

func (m *stateMap) acquire(name string) *State {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	state := m.getUnsafe(name)
	delete(m.white, name)
	delete(m.gray, name)
	return state
}

func (m *stateMap) getUnsafe(name string) *State {
	state := m.white[name]
	if state != nil {
		log.Debugf("State found from white: %s to %s", name, state.Request.URL.String())
		return state
	}
	state = m.gray[name]
	if state != nil {
		log.Debugf("State found from gray: %s to %s", name, state.Request.URL.String())
	} else {
		log.Debugf("State not found: %s", name)
	}
	return state
}

func (m *stateMap) quit() {
	log.Debug("Stopping state map")
	m.stop <- true
}

func (m *stateMap) gc() {
	for {
		select {
		case <-m.stop:
			log.Debug("State map GC stopped")
			break
		case <-m.ticker.C:
			log.Debug("Doing GC for state map")
			m.doGc()
		}
	}
}

func (m *stateMap) doGc() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	log.Debugf("Do GC. white: %d, gray: %d", len(m.white), len(m.gray))
	m.gray = m.white
	m.white = make(map[string]*State)
}
