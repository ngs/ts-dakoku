package app

import (
	"strings"

	"github.com/garyburd/redigo/redis"
)

func (ctx *Context) getUserIDForState(state string) string {
	return strings.Split(ctx.getVariableInHash(ctx.StateStoreKey, state), " ")[0]
}

func (ctx *Context) getTeamIDForState(state string) string {
	return strings.Split(ctx.getVariableInHash(ctx.StateStoreKey, state), " ")[1]
}

func (ctx *Context) storeUserIDInState(teamID string) (string, error) {
	state := ctx.generateState()
	_, err := redis.Bool(ctx.RedisConn.Do("HSET", ctx.StateStoreKey, state, ctx.UserID+" "+teamID))
	return state, err
}

func (ctx *Context) deleteState(state string) error {
	_, err := ctx.RedisConn.Do("HDEL", ctx.StateStoreKey, state)
	return err
}

func (ctx *Context) generateState() string {
	state := ctx.randomString(24)
	exists, _ := redis.Bool(ctx.RedisConn.Do("HEXISTS", ctx.StateStoreKey, state))
	if exists {
		return ctx.generateState()
	}
	return state
}
