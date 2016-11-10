package goru

type Session struct {
	ctx   *Context
	store SessionStore
}

func GetSession(ctx *Context) *Session {
	return &Session{ctx, sessionStore}
}

func (s *Session) Exists(key string) bool {
	return s.store.Exists(s.ctx, "session", key)
}

func (s *Session) Get(key string) (string, bool) {
	return s.store.Get(s.ctx, "session", key)
}

func (s *Session) Set(key, value string) {
	s.store.Set(s.ctx, "session", key, value)
}

func (s *Session) Remove(key string) {
	s.store.Remove(s.ctx, "session", key)
}

func (s *Session) SetFlash(key, value string) {
	s.store.Set(s.ctx, "flash", key, value)
}

func (s *Session) GetFlash(key string) string {
	value, _ := s.store.Get(s.ctx, "flash", key)
	s.store.Remove(s.ctx, "flash", key)
	return value
}

func (s *Session) Clear() {
	s.store.Clear(s.ctx, "session")
	s.store.Clear(s.ctx, "flash")
}

type SessionStore interface {
	Exists(ctx *Context, group, key string) bool
	Get(ctx *Context, group, key string) (string, bool)
	Set(ctx *Context, group, key, value string)
	Remove(ctx *Context, group, key string)
	Clear(ctx *Context, group string)
}

var sessionStore SessionStore = &nullSessionStore{}

func SetSessionStore(store SessionStore) {
	sessionStore = store
}

type nullSessionStore struct {
}

func (s *nullSessionStore) Exists(ctx *Context, group, key string) bool {
	return false
}

func (s *nullSessionStore) Get(ctx *Context, group, key string) (string, bool) {
	return "", false
}

func (s *nullSessionStore) Set(ctx *Context, group, key, value string) {
}

func (s *nullSessionStore) Remove(ctx *Context, group, key string) {
}

func (s *nullSessionStore) Clear(ctx *Context, group string) {
}
