package app

import (
	"os"
	"strings"
	"testing"
)

func createMockApp() *App {
	os.Setenv("STATE_STORE_KEY", "tsdakoku-test:states")
	os.Setenv("OAUTH_TOKEN_STORE_KEY", "tsdakoku-test:oauth_tokens")
	for _, name := range []string{
		"SALESFORCE_CLIENT_SECRET",
		"SALESFORCE_CLIENT_ID",
		"SLACK_VERIFICATION_TOKEN",
		"TEAMSPIRIT_HOST",
	} {
		os.Setenv(name, name+" is set!")
	}
	app, _ := New()
	return app
}

func TestNewApp(t *testing.T) {
	for _, name := range []string{
		"SALESFORCE_CLIENT_SECRET",
		"SALESFORCE_CLIENT_ID",
		"SLACK_VERIFICATION_TOKEN",
		"TEAMSPIRIT_HOST",
		"OAUTH_TOKEN_STORE_KEY",
		"STATE_STORE_KEY",
	} {
		os.Setenv(name, "")
	}
	app, err := New()
	for _, test := range []Test{
		{false, app == nil},
		{"SALESFORCE_CLIENT_SECRET, SALESFORCE_CLIENT_ID, SLACK_VERIFICATION_TOKEN, TEAMSPIRIT_HOST are not configured", err.Error()},
	} {
		test.Compare(t)
	}
	for _, name := range []string{
		"SALESFORCE_CLIENT_SECRET",
		"SALESFORCE_CLIENT_ID",
		"SLACK_VERIFICATION_TOKEN",
		"TEAMSPIRIT_HOST",
	} {
		os.Setenv(name, "ok")
	}
	app, err = New()
	for _, test := range []Test{
		{false, app == nil},
		{true, err == nil},
		{"tsdakoku:states", app.StateStoreKey},
		{"tsdakoku:oauth_tokens", app.TokenStoreKey},
	} {
		test.Compare(t)
	}
	os.Setenv("STATE_STORE_KEY", "tsdakoku-test:states")
	os.Setenv("OAUTH_TOKEN_STORE_KEY", "tsdakoku-test:oauth_tokens")
	app, err = New()
	for _, test := range []Test{
		{false, app == nil},
		{true, err == nil},
		{"tsdakoku-test:states", app.StateStoreKey},
		{"tsdakoku-test:oauth_tokens", app.TokenStoreKey},
	} {
		test.Compare(t)
	}
	origiinalRedisURL := os.Getenv("REDIS_URL")
	os.Setenv("REDIS_URL", "redis://hoge")
	app, err = New()
	for _, test := range []Test{
		{false, app == nil},
		{0, strings.Index(err.Error(), "dial tcp: lookup hoge")},
	} {
		test.Compare(t)
	}
	os.Setenv("REDIS_URL", origiinalRedisURL)
}
