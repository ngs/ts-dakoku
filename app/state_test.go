package app

import (
	"testing"
)

func TestState(t *testing.T) {
	app := createMockApp()
	ctx := app.createContext(nil)
	ctx.UserID = "FOO"
	state, err := ctx.storeUserIDInState()
	for _, test := range []Test{
		{true, err == nil},
		{14, len(state)},
		{"FOO", ctx.getUserIDForState(state)},
	} {
		test.Compare(t)
	}
	err = ctx.deleteState(state)
	for _, test := range []Test{
		{true, err == nil},
		{"", ctx.getUserIDForState(state)},
	} {
		test.Compare(t)
	}
}
