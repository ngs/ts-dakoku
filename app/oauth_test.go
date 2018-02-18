package app

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
	gock "gopkg.in/h2non/gock.v1"
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
		Expiry:       time.Now().Add(-10 * time.Hour),
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

func TestSetAndGetOAuthClient(t *testing.T) {
	defer gock.Off()
	newExpiry := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	oldExpiry := time.Now().Add(-10 * time.Hour).Truncate(time.Second)
	resExpiry, _ := time.Parse("2016-01-02T15:04:05Z", "0001-01-01T00:00:00Z")
	gock.New("https://login.salesforce.com").
		Post("/services/oauth2/token").
		Reply(200).
		JSON(oauth2.Token{
			AccessToken:  "foo2",
			RefreshToken: "bar2",
			TokenType:    "Bearer2",
			Expiry:       resExpiry,
		})
	token := &oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: "bar",
		TokenType:    "Bearer",
		Expiry:       oldExpiry,
	}
	app := createMockApp()
	app.CleanRedis()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)
	ctx := app.createContext(req)
	ctx.UserID = "FOO"
	ctx.TimeoutDuration = 2 * time.Hour
	err := ctx.setAccessToken(token)
	token = ctx.getAccessTokenForUser()
	for _, test := range []Test{
		{false, token == nil},
		{oldExpiry.String(), token.Expiry.String()},
		{"bar", token.RefreshToken},
		{"foo", token.AccessToken},
		{"Bearer", token.TokenType},
		{false, token.Expiry.IsZero()},
		{true, err == nil},
	} {
		test.Compare(t)
	}
	client := ctx.getOAuth2Client()
	token = ctx.getAccessTokenForUser()
	for _, test := range []Test{
		{false, client == nil},
		{newExpiry.String(), token.Expiry.String()},
		{"bar2", token.RefreshToken},
		{"foo2", token.AccessToken},
		{"Bearer2", token.TokenType},
	} {
		test.Compare(t)
	}
}
