package store

import (
	"bytes"
	"github.com/anyufly/gin_session/internal/json"
	"github.com/gorilla/sessions"
	"io"
	"net/http"
)

type ConfigStore interface {
	sessions.Store
	MaxAge(age int)
	MaxLength(length int)
	MinAge(age int)
}

func ReadCookieByRequest(request *http.Request, name string) (string, bool, error) {
	// 首先尝试从cookie中读取
	if cookie, cookieErr := request.Cookie(name); cookieErr == nil {
		// 能进这里说明能读到
		return cookie.Value, true, nil
	}

	// 若cookie中读不到，尝试从url query中读取
	query := request.URL.Query()
	if queryArray, ok := query[name]; ok {
		return queryArray[0], true, nil
	} else {
		// 若还是读不到，尝试从body中读取
		if oldBody := request.Body; oldBody != nil && request.Header.Get("Content-Type") == "application/json" {
			defer func() {
				_ = oldBody.Close()
			}()

			bodyBytes, err := io.ReadAll(oldBody)

			if err != nil {
				return "", false, err
			}

			request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			jsonMap := make(map[string]interface{})
			err = decodeJSON(bytes.NewReader(bodyBytes), &jsonMap)
			if err != nil {
				return "", false, err
			}

			if value, ok := jsonMap[name]; ok {
				return value.(string), true, nil
			}

		}
	}

	return "", false, nil
}

func decodeJSON(r io.Reader, obj any) error {
	decoder := json.NewDecoder(r)

	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}
