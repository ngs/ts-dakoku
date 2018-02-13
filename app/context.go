package app

import (
	"fmt"
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

func (context *Context) SetAccessToken(token string) error {
	fmt.Printf("%v", token)
	return nil // TODO
}
