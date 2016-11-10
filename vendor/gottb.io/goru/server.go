package goru

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var ErrDone = errors.New("server closed")
var ErrTimeout = errors.New("server timed out")

type Server struct {
	srv      *http.Server
	done     bool
	doneChan chan bool
	connMap  map[net.Conn]struct{}
	lock     *sync.Mutex
}

func NewServer(addr string, handler http.Handler) *Server {
	s := &Server{
		srv: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		done:     false,
		doneChan: make(chan bool, 1),
		connMap:  make(map[net.Conn]struct{}),
		lock:     &sync.Mutex{},
	}
	s.srv.ConnState = s.connState
	return s
}

func (srv *Server) ListenAndServe() error {
	addr := srv.srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	signal.Ignore()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	if runtime.GOOS != "windows" {
		signal.Notify(signalChan, syscall.SIGTERM)
	}
	err = srv.srv.Serve(&goruListener{
		ln:         ln.(*net.TCPListener),
		connChan:   make(chan net.Conn, 1),
		errChan:    make(chan error, 1),
		signalChan: signalChan,
		srv:        srv,
	})
	if err == ErrDone {
		return nil
	}
	srv.lock.Lock()
	if len(srv.connMap) == 0 {
		return nil
	}
	srv.lock.Unlock()
	runmode := os.Getenv("RUNMODE")
	if runmode == "development" {
		return nil
	}
	timer := time.NewTimer(10 * time.Second)
	select {
	case <-srv.doneChan:
		return nil
	case <-timer.C:
		return ErrTimeout
	}
}

func (srv *Server) connState(c net.Conn, state http.ConnState) {
	if state == http.StateNew {
		srv.lock.Lock()
		defer srv.lock.Unlock()
		srv.connMap[c] = struct{}{}
	} else if state == http.StateHijacked || state == http.StateClosed {
		srv.lock.Lock()
		defer srv.lock.Unlock()
		delete(srv.connMap, c)
		if srv.done && len(srv.connMap) == 0 {
			srv.doneChan <- true
		}
	}
}

type goruListener struct {
	ln         *net.TCPListener
	connChan   chan net.Conn
	errChan    chan error
	signalChan chan os.Signal
	srv        *Server
}

func (ln *goruListener) Accept() (c net.Conn, err error) {
	go ln.accept()
	select {
	case c := <-ln.connChan:
		return c, nil
	case err := <-ln.errChan:
		return nil, err
	case <-ln.signalChan:
		ln.srv.lock.Lock()
		ln.srv.done = true
		ln.srv.lock.Unlock()
		return nil, ErrDone
	}
}

func (ln *goruListener) accept() {
	tc, err := ln.ln.AcceptTCP()
	if err != nil {
		ln.errChan <- err
	}
	if tc != nil {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
	}
	ln.connChan <- tc
}

func (ln *goruListener) Close() error {
	return ln.ln.Close()
}

func (ln *goruListener) Addr() net.Addr {
	return ln.ln.Addr()
}
