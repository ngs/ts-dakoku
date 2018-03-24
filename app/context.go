package app

import (
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Context in request
type Context struct {
	RedisConn               redis.Conn
	Request                 *http.Request
	SalesforceClientSecret  string
	SalesforceClientID      string
	SlackClientSecret       string
	SlackClientID           string
	UserID                  string
	StateStoreKey           string
	SalesforceTokenStoreKey string
	SlackTokenStoreKey      string
	NotifyChannelStoreKey   string
	TeamSpiritHost          string
	SlackVerificationToken  string
	TimeoutDuration         time.Duration
	TimeTableClient         *timeTableClient
	randomString            func(len int) string
}

func (app *App) createContext(r *http.Request) *Context {
	return &Context{
		RedisConn:               app.RedisConn,
		SalesforceClientID:      app.SalesforceClientID,
		SalesforceClientSecret:  app.SalesforceClientSecret,
		SlackClientID:           app.SlackClientID,
		SlackClientSecret:       app.SlackVerificationToken,
		StateStoreKey:           app.StateStoreKey,
		SalesforceTokenStoreKey: app.SalesforceTokenStoreKey,
		SlackTokenStoreKey:      app.SlackTokenStoreKey,
		NotifyChannelStoreKey:   app.NotifyChannelStoreKey,
		TeamSpiritHost:          app.TeamSpiritHost,
		SlackVerificationToken:  app.SlackVerificationToken,
		TimeoutDuration:         app.TimeoutDuration,
		Request:                 r,
		randomString:            randomString,
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

func (ctx *Context) setVariableInHash(hashKey string, value interface{}) error {
	_, err := redis.Bool(ctx.RedisConn.Do("HSET", hashKey, ctx.UserID, value))
	return err
}
