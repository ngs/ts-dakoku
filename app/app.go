package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	apachelog "github.com/lestrrat/go-apache-logformat"
)

// App main appplication
type App struct {
	Port                    int
	SalesforceClientSecret  string
	SalesforceClientID      string
	SlackClientSecret       string
	SlackClientID           string
	SlackVerificationToken  string
	StateStoreKey           string
	SalesforceTokenStoreKey string
	SlackTokenStoreKey      string
	NotifyChannelStoreKey   string
	TeamSpiritHost          string
	RedisConn               redis.Conn
	TimeoutDuration         time.Duration
}

// New Returns new app
func new() (*App, error) {
	app := &App{}
	salesforceClientSecret := os.Getenv("SALESFORCE_CLIENT_SECRET")
	salesforceClientID := os.Getenv("SALESFORCE_CLIENT_ID")
	slackClientSecret := os.Getenv("SLACK_CLIENT_SECRET")
	slackClientID := os.Getenv("SLACK_CLIENT_ID")
	slackVerificationToken := os.Getenv("SLACK_VERIFICATION_TOKEN")
	teamSpilitHost := os.Getenv("TEAMSPIRIT_HOST")
	var errVars = []string{}
	if salesforceClientSecret == "" {
		errVars = append(errVars, "SALESFORCE_CLIENT_SECRET")
	}
	if salesforceClientID == "" {
		errVars = append(errVars, "SALESFORCE_CLIENT_ID")
	}
	if slackClientSecret == "" {
		errVars = append(errVars, "SLACK_CLIENT_SECRET")
	}
	if slackClientID == "" {
		errVars = append(errVars, "SLACK_CLIENT_ID")
	}
	if slackVerificationToken == "" {
		errVars = append(errVars, "SLACK_VERIFICATION_TOKEN")
	}
	if teamSpilitHost == "" {
		errVars = append(errVars, "TEAMSPIRIT_HOST")
	}
	if len(errVars) > 0 {
		return app, fmt.Errorf("%s are not configured", strings.Join(errVars, ", "))
	}

	if k := os.Getenv("STATE_STORE_KEY"); k != "" {
		app.StateStoreKey = k
	} else {
		app.StateStoreKey = "tsdakoku:states"
	}

	if k := os.Getenv("OAUTH_TOKEN_STORE_KEY"); k != "" {
		app.SalesforceTokenStoreKey = k
	} else {
		app.SalesforceTokenStoreKey = "tsdakoku:oauth_tokens"
	}

	if k := os.Getenv("SLACK_TOKEN_STORE_KEY"); k != "" {
		app.SlackTokenStoreKey = k
	} else {
		app.SlackTokenStoreKey = "tsdakoku:slack_tokens"
	}

	if k := os.Getenv("SLACK_NOTIFY_CHANNEL_STORE_KEY"); k != "" {
		app.NotifyChannelStoreKey = k
	} else {
		app.NotifyChannelStoreKey = "tsdakoku:notify_channels"
	}

	duration, _ := strconv.Atoi(os.Getenv("SALESFORCE_TIMEOUT_MINUTES"))
	if duration > 0 {
		app.TimeoutDuration = time.Duration(duration) * time.Minute
	} else {
		app.TimeoutDuration = time.Hour
	}

	app.SalesforceClientID = salesforceClientID
	app.SalesforceClientSecret = salesforceClientSecret
	app.SlackClientID = slackClientID
	app.SlackClientSecret = slackClientSecret
	app.SlackVerificationToken = slackVerificationToken
	app.TeamSpiritHost = teamSpilitHost
	if err := app.setupRedis(); err != nil {
		return app, err
	}
	return app, nil
}

// Run starts web server
func Run() (*App, error) {
	app, err := new()
	if err != nil {
		return app, err
	}
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if !(port > 0) {
		port = 8000
	}
	app.Port = port
	router := app.setupRouter()
	fmt.Println("Listeninng on 0.0.0.0:" + strconv.Itoa(port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), apachelog.CombinedLog.Wrap(router, os.Stderr)))
	return app, nil
}
