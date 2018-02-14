package app

import (
	"testing"
)

func TestState(t *testing.T) {
	app := createMockApp()
	ctx := app.CreateContext(nil)
	ctx.UserID = "FOO"
	state, err := ctx.StoreUserIDInState()
	for _, test := range []Test{
		{true, err == nil},
		{14, len(state)},
		{"FOO", ctx.GetUserIDForState(state)},
	} {
		test.Compare(t)
	}
	err = ctx.DeleteState(state)
	for _, test := range []Test{
		{true, err == nil},
		{"", ctx.GetUserIDForState(state)},
	} {
		test.Compare(t)
	}
}
