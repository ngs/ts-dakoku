package app

import (
	"net/http"
	"testing"

	"golang.org/x/oauth2"
)

func TestGetOAuthCallbackURL(t *testing.T) {
	app := createMockApp()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)
	ctx := app.createContext(req)
	Test{"https://example.com/oauth/callback", ctx.getOAuthCallbackURL()}.Compare(t)
}

func TestGetAuthenticateURL(t *testing.T) {
	app := createMockApp()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)
	ctx := app.createContext(req)
	Test{"https://example.com/oauth/authenticate/foo", ctx.getAuthenticateURL("foo")}.Compare(t)
}

func TestSetAndGetAccessToken(t *testing.T) {
	token := &oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: "bar",
		TokenType:    "Bearer",
	}
	app := createMockApp()
	app.CleanRedis()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)
	ctx := app.createContext(req)
	ctx.UserID = "FOO"
	err := ctx.setAccessToken(token)
	Test{false, err != nil}.Compare(t)
	token = ctx.getAccessTokenForUser()
	for _, test := range []Test{
		{"foo", token.AccessToken},
		{"bar", token.RefreshToken},
		{"Bearer", token.TokenType},
	} {
		test.Compare(t)
	}
	ctx = app.createContext(req)
	ctx.UserID = "BAR"
	token = ctx.getAccessTokenForUser()
	Test{true, token == nil}.Compare(t)
}
