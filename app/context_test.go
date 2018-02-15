package app

import (
	"net/http"
	"testing"
)

func TestCreateContext(t *testing.T) {
	app := createMockApp()
	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	ctx := app.CreateContext(req)

	for _, test := range []Test{
		{false, ctx.RedisConn == nil},
		{"SALESFORCE_CLIENT_ID is set!", ctx.ClientID},
		{"SALESFORCE_CLIENT_SECRET is set!", ctx.ClientSecret},
		{"tsdakoku-test:states", ctx.StateStoreKey},
		{"tsdakoku-test:oauth_tokens", ctx.TokenStoreKey},
		{"teamspirit-1234.cloudforce.test", ctx.TeamSpiritHost},
		{"SLACK_VERIFICATION_TOKEN is set!", ctx.SlackVerificationToken},
		{req, ctx.Request},
	} {
		test.Compare(t)
	}
}
