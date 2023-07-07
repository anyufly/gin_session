package store

import (
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
)

type CookieStore struct {
	store *sessions.CookieStore
}

func NewCookieStore(keyPairs ...[]byte) *CookieStore {
	cs := &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
	}

	cs.MaxAge(cs.Options.MaxAge)

	return &CookieStore{
		store: cs,
	}

}

func (cs *CookieStore) MaxAge(age int) {
	cs.store.MaxAge(age)
}

func (cs *CookieStore) MaxLength(length int) {
	for _, c := range cs.store.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MaxLength(length)
		}
	}
}

func (cs *CookieStore) MinAge(age int) {
	for _, c := range cs.store.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MinAge(age)
		}
	}
}

func (cs *CookieStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return cs.store.Get(r, name)
}

func (cs *CookieStore) New(r *http.Request, name string) (*sessions.Session, error) {
	s := sessions.NewSession(cs, name)
	opts := *cs.store.Options
	s.Options = &opts
	s.IsNew = true

	content, read, err := ReadCookieByRequest(r, name)

	if err != nil {
		return s, err
	}

	if !read {
		return s, nil
	}

	err = securecookie.DecodeMulti(name, content, &s.Values,
		cs.store.Codecs...)
	if err == nil {
		s.IsNew = false
	}

	return s, err
}

func (cs *CookieStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	return cs.store.Save(r, w, s)
}
