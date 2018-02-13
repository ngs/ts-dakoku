package app

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

func (app *App) GetOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		Scopes:       []string{"full"},
		RedirectURL:  os.Getenv("SF_OAUTH_CALLBACK_URL"),
		Endpoint: oauth2.Endpoint{
			// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_understanding_oauth_endpoints.htm
			AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
			TokenURL: "https://login.salesforce.com/services/oauth2/token",
		},
	}
}

func (app *App) GetAccessToken(code string, state string) (string, error) {
	config := app.GetOAuth2Config()
	fmt.Printf("state: %+v\n", state)
	ctx := context.TODO()
	t, err := config.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	return t.AccessToken, nil
}

func (context *Context) GetOAuth2Client() *http.Client {
	token := context.GetAccessToken()
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
