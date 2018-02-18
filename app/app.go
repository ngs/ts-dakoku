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
	Port                   int
	ClientSecret           string
	ClientID               string
	SlackVerificationToken string
	StateStoreKey          string
	TokenStoreKey          string
	TeamSpiritHost         string
	RedisConn              redis.Conn
	TimeoutDuration        time.Duration
}

// New Returns new app
func new() (*App, error) {
	app := &App{}
	clientSecret := os.Getenv("SALESFORCE_CLIENT_SECRET")
	clientID := os.Getenv("SALESFORCE_CLIENT_ID")
	slackVerificationToken := os.Getenv("SLACK_VERIFICATION_TOKEN")
	teamSpilitHost := os.Getenv("TEAMSPIRIT_HOST")
	var errVars = []string{}
	if clientSecret == "" {
		errVars = append(errVars, "SALESFORCE_CLIENT_SECRET")
	}
	if clientID == "" {
		errVars = append(errVars, "SALESFORCE_CLIENT_ID")
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
		app.TokenStoreKey = k
	} else {
		app.TokenStoreKey = "tsdakoku:oauth_tokens"
	}

	duration, _ := strconv.Atoi(os.Getenv("SALESFORCE_TIMEOUT_MINUTES"))
	if duration > 0 {
		app.TimeoutDuration = time.Duration(duration) * time.Minute
	} else {
		app.TimeoutDuration = time.Hour
	}

	app.ClientID = clientID
	app.ClientSecret = clientSecret
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
