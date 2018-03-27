package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

func (ctx *Context) getSalesforceOAuthCallbackURL() string {
	return "https://" + ctx.Request.Host + "/oauth/salesforce/callback"
}

func (ctx *Context) getSalesforceAuthenticateURL(state string) string {
	return "https://" + ctx.Request.Host + "/oauth/salesforce/authenticate/" + state
}

func (ctx *Context) getSlackOAuthCallbackURL() string {
	return "https://" + ctx.Request.Host + "/oauth/slack/callback"
}

func (ctx *Context) getSlackAuthenticateURL(teamID, state string) string {
	return "https://" + ctx.Request.Host + "/oauth/slack/authenticate/" + teamID + "/" + state
}

func (ctx *Context) setSalesforceAccessToken(token *oauth2.Token) error {
	if ctx.UserID == "" {
		return errors.New("UserID is not set")
	}
	// SalesForce always returns zero-expiry, but it expires.
	if token.Expiry.IsZero() {
		token.Expiry = time.Now().Add(ctx.TimeoutDuration).Truncate(time.Second)
	}
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return ctx.setVariableInHash(ctx.SalesforceTokenStoreKey, tokenJSON)
}

func (ctx *Context) setSlackAccessToken(token string) error {
	if ctx.UserID == "" {
		return errors.New("UserID is not set")
	}
	return ctx.setVariableInHash(ctx.SlackTokenStoreKey, token)
}

func (ctx *Context) getSalesforceOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     ctx.SalesforceClientID,
		ClientSecret: ctx.SalesforceClientSecret,
		Scopes:       []string{},
		RedirectURL:  ctx.getSalesforceOAuthCallbackURL(),
		Endpoint: oauth2.Endpoint{
			// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_understanding_oauth_endpoints.htm
			AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
			TokenURL: "https://login.salesforce.com/services/oauth2/token",
		},
	}
}

func (ctx *Context) getSalesforceAccessTokenForUser() *oauth2.Token {
	if ctx.UserID == "" {
		return nil
	}
	tokenJSON := ctx.getVariableInHash(ctx.SalesforceTokenStoreKey, ctx.UserID)
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil
	}
	return &token
}

func (ctx *Context) getSlackAccessTokenForUser() string {
	return ctx.getVariableInHash(ctx.SlackTokenStoreKey, ctx.UserID)
}

func (ctx *Context) getSlackNotifyChannelForUser() string {
	return ctx.getVariableInHash(ctx.NotifyChannelStoreKey, ctx.UserID)
}

func (ctx *Context) getSalesforceAccessToken(code string, state string) (*oauth2.Token, error) {
	config := ctx.getSalesforceOAuth2Config()
	t, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (ctx *Context) getSalesforceOAuth2Client() *http.Client {
	token := ctx.getSalesforceAccessTokenForUser()
	if token == nil {
		return nil
	}
	src := ctx.getSalesforceOAuth2Config().TokenSource(context.TODO(), token)
	ts := oauth2.ReuseTokenSource(token, src)
	if token, _ := ts.Token(); token != nil {
		ctx.setSalesforceAccessToken(token)
	}
	return oauth2.NewClient(oauth2.NoContext, ts)
}
