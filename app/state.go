package app

import (
	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

// State state for authentication
type State struct {
	UserID      string `json:"u,omitempty"`
	TeamID      string `json:"t,omitempty"`
	ResponseURL string `json:"r,omitempty"`
}

func (ctx *Context) getState(state string) *State {
	res, err := ctx.RedisConn.Do("HGET", ctx.StateStoreKey, state)
	if err != nil {
		return nil
	}
	if data, ok := res.([]byte); ok {
		var res State
		json.Unmarshal(data, &res)
		return &res
	}
	return nil
}

func (ctx *Context) storeState(state State) (string, error) {
	state.UserID = ctx.UserID
	stateKey := ctx.generateState()
	jsonData, _ := json.Marshal(state)
	_, err := redis.Bool(ctx.RedisConn.Do("HSET", ctx.StateStoreKey, stateKey, string(jsonData)))
	return stateKey, err
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
