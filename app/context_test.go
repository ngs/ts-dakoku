package app

import (
	"net/http"
	"testing"
)

func TestCreateContext(t *testing.T) {
	app := createMockApp()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)
	ctx := app.createContext(req)

	for _, test := range []Test{
		{false, ctx.RedisConn == nil},
		{"SALESFORCE_CLIENT_ID is set!", ctx.SalesforceClientID},
		{"SALESFORCE_CLIENT_SECRET is set!", ctx.SalesforceClientSecret},
		{"tsdakoku-test:states", ctx.StateStoreKey},
		{"tsdakoku-test:oauth_tokens", ctx.SalesforceTokenStoreKey},
		{"teamspirit-1234.cloudforce.test", ctx.TeamSpiritHost},
		{"SLACK_VERIFICATION_TOKEN is set!", ctx.SlackVerificationToken},
		{req, ctx.Request},
	} {
		test.Compare(t)
	}
}
