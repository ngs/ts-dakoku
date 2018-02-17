package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/oauth2"
)

func (ctx *Context) getOAuthCallbackURL() string {
	return "https://" + ctx.Request.Host + "/oauth/callback"
}

func (ctx *Context) getAuthenticateURL(state string) string {
	return "https://" + ctx.Request.Host + "/oauth/authenticate/" + state
}

func (ctx *Context) setAccessToken(token *oauth2.Token) error {
	if ctx.UserID == "" {
		return errors.New("UserID is not set")
	}
	// SalesForce always returns zero-expiry, but it expires.
	if token.Expiry.IsZero() {
		token.Expiry = time.Now().Add(time.Hour).Truncate(time.Second)
	}
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}
	_, err = redis.Bool(ctx.RedisConn.Do("HSET", ctx.TokenStoreKey, ctx.UserID, tokenJSON))
	return err
}

func (ctx *Context) getOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     ctx.ClientID,
		ClientSecret: ctx.ClientSecret,
		Scopes:       []string{},
		RedirectURL:  ctx.getOAuthCallbackURL(),
		Endpoint: oauth2.Endpoint{
			// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_understanding_oauth_endpoints.htm
			AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
			TokenURL: "https://login.salesforce.com/services/oauth2/token",
		},
	}
}

func (ctx *Context) getAccessTokenForUser() *oauth2.Token {
	if ctx.UserID == "" {
		return nil
	}
	tokenJSON := ctx.getVariableInHash(ctx.TokenStoreKey, ctx.UserID)
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil
	}
	return &token
}

func (ctx *Context) getAccessToken(code string, state string) (*oauth2.Token, error) {
	config := ctx.getOAuth2Config()
	t, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (ctx *Context) getOAuth2Client() *http.Client {
	token := ctx.getAccessTokenForUser()
	if token == nil {
		return nil
	}
	src := ctx.getOAuth2Config().TokenSource(context.TODO(), token)
	ts := oauth2.ReuseTokenSource(token, src)
	if token, _ := ts.Token(); token != nil {
		ctx.setAccessToken(token)
	}
	return oauth2.NewClient(oauth2.NoContext, ts)
}
