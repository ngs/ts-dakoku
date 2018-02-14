package app

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/garyburd/redigo/redis"
)

type Context struct {
	RedisConn              redis.Conn
	Request                *http.Request
	ClientSecret           string
	ClientID               string
	UserID                 string
	StateStoreKey          string
	TokenStoreKey          string
	TeamSpiritHost         string
	SlackVerificationToken string
}

func (app *App) CreateContext(r *http.Request) *Context {
	ctx := &Context{
		RedisConn:              app.RedisConn,
		ClientID:               app.ClientID,
		ClientSecret:           app.ClientSecret,
		StateStoreKey:          app.StateStoreKey,
		TokenStoreKey:          app.TokenStoreKey,
		TeamSpiritHost:         app.TeamSpiritHost,
		SlackVerificationToken: app.SlackVerificationToken,
		Request:                r,
	}
	return ctx
}

func (ctx *Context) GetOAuthCallbackURL() string {
	return "https://" + ctx.Request.Host + "/oauth/callback"
}

func (ctx *Context) GetAuthenticateURL(state string) string {
	return "https://" + ctx.Request.Host + "/oauth/authenticate/" + state
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

func (ctx *Context) SetAccessToken(token *oauth2.Token) error {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}
	_, err = redis.Bool(ctx.RedisConn.Do("HSET", ctx.TokenStoreKey, ctx.UserID, tokenJSON))
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

func (ctx *Context) GetAccessTokenForUser() *oauth2.Token {
	tokenJSON := ctx.getVariableInHash(ctx.TokenStoreKey, ctx.UserID)
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil
	}
	return &token
}

func (ctx *Context) GetAccessToken(code string, state string) (*oauth2.Token, error) {
	config := ctx.GetOAuth2Config()
	t, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (ctx *Context) GetOAuth2Client() *http.Client {
	token := ctx.GetAccessTokenForUser()
	if token == nil {
		return nil
	}
	src := ctx.GetOAuth2Config().TokenSource(context.TODO(), token)
	ts := oauth2.ReuseTokenSource(token, src)
	return oauth2.NewClient(oauth2.NoContext, ts)
}
