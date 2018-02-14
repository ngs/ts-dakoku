package app

import "github.com/garyburd/redigo/redis"

func (ctx *Context) GetUserIDForState(state string) string {
	return ctx.getVariableInHash(ctx.StateStoreKey, state)
}

func (ctx *Context) StoreUserIDInState() (string, error) {
	state := ctx.generateState()
	_, err := redis.Bool(ctx.RedisConn.Do("HSET", ctx.StateStoreKey, state, ctx.UserID))
	return state, err
}

func (ctx *Context) DeleteState(state string) error {
	_, err := ctx.RedisConn.Do("HDEL", ctx.StateStoreKey, state)
	return err
}

func (ctx *Context) generateState() string {
	state := RandomString(24)
	exists, _ := redis.Bool(ctx.RedisConn.Do("HEXISTS", ctx.StateStoreKey, state))
	if exists {
		return ctx.generateState()
	}
	return state
}
