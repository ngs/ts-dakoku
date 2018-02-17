package app

import (
	"net/http"

	"github.com/garyburd/redigo/redis"
)

// Context in request
type Context struct {
	RedisConn              redis.Conn
	Request                *http.Request
	ClientSecret           string
	ClientID               string
	UserID                 string
	StateStoreKey          string
	TokenStoreKey          string
	TeamSpiritHost         string
	SlackVerificationToken string
	TimeTableClient        *timeTableClient
}

func (app *App) createContext(r *http.Request) *Context {
	return &Context{
		RedisConn:              app.RedisConn,
		ClientID:               app.ClientID,
		ClientSecret:           app.ClientSecret,
		StateStoreKey:          app.StateStoreKey,
		TokenStoreKey:          app.TokenStoreKey,
		TeamSpiritHost:         app.TeamSpiritHost,
		SlackVerificationToken: app.SlackVerificationToken,
		Request:                r,
	}
}

func (ctx *Context) getVariableInHash(hashKey string, key string) string {
	res, err := ctx.RedisConn.Do("HGET", hashKey, key)
	if err != nil {
		return ""
	}
	if data, ok := res.([]byte); ok {
		return string(data)
	}
	return ""
}
