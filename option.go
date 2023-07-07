package session

import (
	"github.com/anyufly/gin_session/store"
	"github.com/gorilla/sessions"
	"net/http"
)

type Option interface {
	Apply(s *RequestSession)
}

type MaxAge int

func (m MaxAge) Apply(s *RequestSession) {
	if configStore, ok := s.Store.(store.ConfigStore); ok {
		configStore.MaxAge(int(m))
	}

}

type MaxLength int

func (m MaxLength) Apply(s *RequestSession) {
	if configStore, ok := s.Store.(store.ConfigStore); ok {
		configStore.MaxLength(int(m))
	}

}

type MinAge int

func (m MinAge) Apply(s *RequestSession) {
	if configStore, ok := s.Store.(store.ConfigStore); ok {
		configStore.MinAge(int(m))
	}
}

type Session interface {
	Get(name string) (*sessions.Session, error)
	Save(w http.ResponseWriter, ss *sessions.Session) error
}
