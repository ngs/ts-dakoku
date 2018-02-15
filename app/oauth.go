package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/oauth2"
)

func (ctx *Context) GetOAuthCallbackURL() string {
	return "https://" + ctx.Request.Host + "/oauth/callback"
}

func (ctx *Context) GetAuthenticateURL(state string) string {
	return "https://" + ctx.Request.Host + "/oauth/authenticate/" + state
}

func (ctx *Context) SetAccessToken(token *oauth2.Token) error {
	if ctx.UserID == "" {
		return errors.New("UserID is not set")
	}
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
