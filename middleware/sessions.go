package middleware

import (
	"github.com/anyufly/gin_session"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func Sessions(store sessions.Store, sessionKey string, opts ...session.Option) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s := &session.RequestSession{
			Request: ctx.Request,
			Store:   store,
		}

		for _, opt := range opts {
			opt.Apply(s)
		}

		ctx.Set(sessionKey, s)
		ctx.Next()
	}
}

func GetRequestSession(ctx *gin.Context, sessionKey string) session.Session {
	return ctx.MustGet(sessionKey).(session.Session)
}
