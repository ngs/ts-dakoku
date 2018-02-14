package app

import (
	"net/http"
	"testing"

	"golang.org/x/oauth2"
)

func TestGetOAuthCallbackURL(t *testing.T) {
	app := createMockApp()
	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	ctx := app.CreateContext(req)
	Test{"https://example.com/oauth/callback", ctx.GetOAuthCallbackURL()}.Compare(t)
}

func TestGetAuthenticateURL(t *testing.T) {
	app := createMockApp()
	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	ctx := app.CreateContext(req)
	Test{"https://example.com/oauth/authenticate/foo", ctx.GetAuthenticateURL("foo")}.Compare(t)
}

func TestSetAndGetAccessToken(t *testing.T) {
	token := &oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: "bar",
		TokenType:    "Bearer",
	}
	app := createMockApp()
	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	ctx := app.CreateContext(req)
	ctx.UserID = "FOO"
	err := ctx.SetAccessToken(token)
	Test{false, err != nil}.Compare(t)
	token = ctx.GetAccessTokenForUser()
	for _, test := range []Test{
		{"foo", token.AccessToken},
		{"bar", token.RefreshToken},
		{"Bearer", token.TokenType},
	} {
		test.Compare(t)
	}
	ctx.UserID = "BAR"
	token = ctx.GetAccessTokenForUser()
	Test{true, token == nil}.Compare(t)
}
