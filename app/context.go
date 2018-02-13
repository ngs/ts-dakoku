package app

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/garyburd/redigo/redis"
)

type Context struct {
	RedisConn    redis.Conn
	Request      *http.Request
	ClientSecret string
	ClientID     string
	UserID       string
}

func (app *App) CreateContext(r *http.Request) *Context {
	ctx := &Context{
		RedisConn:    app.RedisConn,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		Request:      r,
	}
	return ctx
}

func (ctx *Context) GetOAuthCallbackURL() string {
	return "https://" + ctx.Request.Host + "/oauth/callback"
}

func (ctx *Context) GetAccessTokenForUser() string {
	return "" // TODO
}

func (ctx *Context) SetAccessToken(token string) error {
	fmt.Printf("%v", token)
	return nil // TODO
}

func (ctx *Context) GetOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     ctx.ClientID,
		ClientSecret: ctx.ClientSecret,
		Scopes:       []string{"full"},
		RedirectURL:  ctx.GetOAuthCallbackURL(),
		Endpoint: oauth2.Endpoint{
			// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_understanding_oauth_endpoints.htm
			AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
			TokenURL: "https://login.salesforce.com/services/oauth2/token",
		},
	}
}

func (ctx *Context) GetAccessToken(code string, state string) (string, error) {
	config := ctx.GetOAuth2Config()
	fmt.Printf("state: %+v\n", state)
	t, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return "", err
	}
	return t.AccessToken, nil
}

func (ctx *Context) GetOAuth2Client() *http.Client {
	token := ctx.GetAccessTokenForUser()
	if token == "" {
		return nil
	}
	return GetOAuth2ClientForToken(token)
}

func GetOAuth2ClientForToken(token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(oauth2.NoContext, ts)
}
