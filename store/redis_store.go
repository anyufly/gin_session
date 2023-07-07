package store

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

const defaultPrefix = "_session"

type RedisStore struct {
	rdb     redis.UniversalClient
	Options *sessions.Options
	Codecs  []securecookie.Codec
	prefix  string
}

func NewRedisStore(rdb redis.UniversalClient, keyPairs ...[]byte) *RedisStore {
	if rdb == nil {
		panic("rdb is nil")
	}

	rs := &RedisStore{
		rdb: rdb,
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
	}

	rs.MaxAge(rs.Options.MaxAge)
	return rs
}

func (rs *RedisStore) SetPrefix(prefix string) {
	rs.prefix = prefix
}

func (rs *RedisStore) getPrefix() string {
	if rs.prefix == "" {
		return defaultPrefix
	}

	return rs.prefix
}

func (rs *RedisStore) buildSessionKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", rs.getPrefix(), sessionID)
}

func (rs *RedisStore) MaxAge(age int) {
	rs.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range rs.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (rs *RedisStore) MaxLength(length int) {
	for _, c := range rs.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MaxLength(length)
		}
	}
}

func (rs *RedisStore) MinAge(age int) {
	for _, c := range rs.Codecs {
		if codec, ok := c.(*securecookie.SecureCookie); ok {
			codec.MinAge(age)
		}
	}
}

func (rs *RedisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(rs, name)
}

func (rs *RedisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	s := sessions.NewSession(rs, name)
	opts := *rs.Options
	s.Options = &opts
	s.IsNew = true

	content, read, err := ReadCookieByRequest(r, name)

	if err != nil {
		return s, err
	}

	if !read {
		return s, nil
	}

	err = securecookie.DecodeMulti(name, content, &s.ID, rs.Codecs...)

	if err == nil {
		err = rs.load(s)
		if err == nil {
			s.IsNew = false
		} else if err == redis.Nil {
			err = nil
		}
	}

	return s, err

}

func (rs *RedisStore) load(session *sessions.Session) error {

	if cmd := rs.rdb.Get(context.Background(), rs.buildSessionKey(session.ID)); cmd.Err() != nil {
		return cmd.Err()
	} else {
		result := cmd.Val()
		if err := securecookie.DecodeMulti(session.Name(), result,
			&session.Values, rs.Codecs...); err != nil {
			return err
		}
	}

	return nil
}

func (rs *RedisStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	if s.Options.MaxAge <= 0 {
		if err := rs.delete(s); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(s.Name(), "", s.Options))
		return nil
	}

	if s.ID == "" {
		s.ID = uuid.NewString()
	}

	if err := rs.save(s); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(s.Name(), s.ID,
		rs.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(s.Name(), encoded, s.Options))
	return nil
}

func (rs *RedisStore) save(s *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(s.Name(), s.Values,
		rs.Codecs...)
	if err != nil {
		return err
	}
	if cmd := rs.rdb.SetEx(context.Background(),
		rs.buildSessionKey(s.ID),
		encoded,
		time.Duration(s.Options.MaxAge)*time.Second); cmd.Err() != nil {
		return err
	}

	return nil
}

func (rs *RedisStore) delete(s *sessions.Session) error {
	if cmd := rs.rdb.Del(context.Background(), rs.buildSessionKey(s.ID)); cmd.Err() != nil {
		return cmd.Err()
	}

	return nil
}
