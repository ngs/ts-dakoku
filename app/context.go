package app

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/garyburd/redigo/redis"
)

type Context struct {
	RedisConn           redis.Conn
	Request             *http.Request
	ClientSecret        string
	ClientID            string
	UserID              string
	StateStoreKey       string
	AccessTokenStoreKey string
}

func (app *App) CreateContext(r *http.Request) *Context {
	ctx := &Context{
		RedisConn:           app.RedisConn,
		ClientID:            app.ClientID,
		ClientSecret:        app.ClientSecret,
		StateStoreKey:       app.StateStoreKey,
		AccessTokenStoreKey: app.AccessTokenStoreKey,
		Request:             r,
	}
	return ctx
}

func (ctx *Context) GetOAuthCallbackURL() string {
	return "https://" + ctx.Request.Host + "/oauth/callback"
}

func (ctx *Context) GetAuthenticateURL(state string) string {
	return "https://" + ctx.Request.Host + "/oauth/authenticate/" + state
}
func (ctx *Context) GetAccessTokenForUser() string {
	return ctx.getVariableInHash(ctx.AccessTokenStoreKey, ctx.UserID)
}

func (ctx *Context) GetUserIDForState(state string) string {
	return ctx.getVariableInHash(ctx.StateStoreKey, state)
}

func (ctx *Context) getVariableInHash(hashKey string, key string) string {
	res, err := ctx.RedisConn.Do("HGET", hashKey, key)
	if err != nil {
		return ""
	}
	if data, ok := res.([]byte); ok {
		return string(data)
	}
	return ""
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

func (ctx *Context) SetAccessToken(token string) error {
	_, err := redis.Bool(ctx.RedisConn.Do("HSET", ctx.AccessTokenStoreKey, ctx.UserID, token))
	return err
}

func (ctx *Context) GetOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     ctx.ClientID,
		ClientSecret: ctx.ClientSecret,
		Scopes:       []string{},
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
