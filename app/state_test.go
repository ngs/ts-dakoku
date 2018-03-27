package app

import (
	"fmt"
	"testing"
)

func TestState(t *testing.T) {
	app := createMockApp()
	app.CleanRedis()
	ctx := app.createContext(nil)
	ctx.UserID = "FOO"
	var callCount = 0
	ctx.randomString = func(len int) string {
		callCount++
		return fmt.Sprintf("random-%d-%d", len, callCount)
	}
	ctx.RedisConn.Do("HSET", ctx.StateStoreKey, "random-24-1", "BAR")
	state, err := ctx.storeState(State{TeamID: "T123456", UserID: "FOO", ResponseURL: "http://foo.com/bar"})
	for _, test := range []Test{
		{true, err == nil},
		{"random-24-2", state},
		{"FOO", ctx.getState(state).UserID},
		{"T123456", ctx.getState(state).TeamID},
	} {
		test.Compare(t)
	}
	err = ctx.deleteState(state)
	for _, test := range []Test{
		{true, err == nil},
		{true, ctx.getState(state) == nil},
	} {
		test.Compare(t)
	}
}
