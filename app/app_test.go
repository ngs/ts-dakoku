package app

import (
	"os"
	"strings"
	"testing"
	"time"
)

func (app *App) CleanRedis() {
	app.RedisConn.Do("DEL", app.SalesforceTokenStoreKey)
	app.RedisConn.Do("DEL", app.StateStoreKey)
}

func createMockApp() *App {
	os.Setenv("STATE_STORE_KEY", "tsdakoku-test:states")
	os.Setenv("OAUTH_TOKEN_STORE_KEY", "tsdakoku-test:oauth_tokens")
	for _, name := range []string{
		"SALESFORCE_CLIENT_SECRET",
		"SALESFORCE_CLIENT_ID",
		"SLACK_VERIFICATION_TOKEN",
	} {
		os.Setenv(name, name+" is set!")
	}
	os.Setenv("TEAMSPIRIT_HOST", "teamspirit-1234.cloudforce.test")
	app, _ := new()
	return app
}

func TestNewApp(t *testing.T) {
	for _, name := range []string{
		"SALESFORCE_CLIENT_SECRET",
		"SALESFORCE_CLIENT_ID",
		"SLACK_CLIENT_SECRET",
		"SLACK_CLIENT_ID",
		"SLACK_VERIFICATION_TOKEN",
		"TEAMSPIRIT_HOST",
		"OAUTH_TOKEN_STORE_KEY",
		"STATE_STORE_KEY",
	} {
		os.Setenv(name, "")
	}
	app, err := new()
	for _, test := range []Test{
		{false, app == nil},
		{"SALESFORCE_CLIENT_SECRET, SALESFORCE_CLIENT_ID, SLACK_CLIENT_SECRET, SLACK_CLIENT_ID, SLACK_VERIFICATION_TOKEN, TEAMSPIRIT_HOST are not configured", err.Error()},
	} {
		test.Compare(t)
	}
	for _, name := range []string{
		"SALESFORCE_CLIENT_SECRET",
		"SALESFORCE_CLIENT_ID",
		"SLACK_CLIENT_SECRET",
		"SLACK_CLIENT_ID",
		"SLACK_VERIFICATION_TOKEN",
		"TEAMSPIRIT_HOST",
	} {
		os.Setenv(name, "ok")
	}
	app, err = new()
	for _, test := range []Test{
		{false, app == nil},
		{true, err == nil},
		{"tsdakoku:states", app.StateStoreKey},
		{"tsdakoku:oauth_tokens", app.SalesforceTokenStoreKey},
		{time.Hour, app.TimeoutDuration},
	} {
		test.Compare(t)
	}
	os.Setenv("STATE_STORE_KEY", "tsdakoku-test:states")
	os.Setenv("OAUTH_TOKEN_STORE_KEY", "tsdakoku-test:oauth_tokens")
	os.Setenv("SLACK_TOKEN_STORE_KEY", "tsdakoku-test:slack_tokens")
	os.Setenv("SALESFORCE_TIMEOUT_MINUTES", "20")
	app, err = new()
	for _, test := range []Test{
		{false, app == nil},
		{true, err == nil},
		{"tsdakoku-test:states", app.StateStoreKey},
		{"tsdakoku-test:oauth_tokens", app.SalesforceTokenStoreKey},
		{"tsdakoku-test:slack_tokens", app.SlackTokenStoreKey},
		{20 * time.Minute, app.TimeoutDuration},
	} {
		test.Compare(t)
	}
	os.Setenv("SALESFORCE_TIMEOUT_MINUTES", "100hoge")

	app, err = new()
	for _, test := range []Test{
		{time.Hour, app.TimeoutDuration},
	} {
		test.Compare(t)
	}

	origiinalRedisURL := os.Getenv("REDIS_URL")
	os.Setenv("REDIS_URL", "redis://hoge")
	os.Setenv("SALESFORCE_TIMEOUT_MINUTES", "")
	app, err = new()
	for _, test := range []Test{
		{false, app == nil},
		{0, strings.Index(err.Error(), "dial tcp: lookup hoge")},
	} {
		test.Compare(t)
	}
	os.Setenv("REDIS_URL", origiinalRedisURL)
}

func TestRun(t *testing.T) {
	go func() {
		_, err := Run()
		if err != nil {
			t.Fatal(err.Error())
		}
	}()
	time.Sleep(time.Second)
	os.Setenv("SALESFORCE_CLIENT_SECRET", "")
	go func() {
		_, err := Run()
		Test{"SALESFORCE_CLIENT_SECRET are not configured", err.Error()}.Compare(t)
	}()
	time.Sleep(time.Second)
}
