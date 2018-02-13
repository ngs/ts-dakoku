package app

import (
	"net/http"

	"github.com/garyburd/redigo/redis"
)

type Context struct {
	RedisConn redis.Conn
	Request   *http.Request
}

func (app *App) CreateContext(r *http.Request) *Context {
	ctx := &Context{
		RedisConn: app.RedisConn,
		Request:   r,
	}
	return ctx
}

func (context *Context) GetAccessToken() string {
	return "" // TODO
}

func (context *Context) SetAccessToken(token string, w http.ResponseWriter) error {
	return nil // TODO
}
