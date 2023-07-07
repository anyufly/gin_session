package session

import (
	"github.com/gorilla/sessions"
	"net/http"
)

type RequestSession struct {
	Request *http.Request
	Store   sessions.Store
}

func (s *RequestSession) Get(name string) (*sessions.Session, error) {
	return s.Store.Get(s.Request, name)
}

func (s *RequestSession) Save(w http.ResponseWriter, ss *sessions.Session) error {
	return s.Store.Save(s.Request, w, ss)
}
