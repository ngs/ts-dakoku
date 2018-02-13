package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
	apachelog "github.com/lestrrat/go-apache-logformat"
)

type App struct {
	Port                   int
	ClientSecret           string
	ClientID               string
	SlackVerificationToken string
	StateStoreKey          string
	AccessTokenStoreKey    string
	TeamSpiritHost         string
	RedisConn              redis.Conn
}

func New() (*App, error) {
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
		return app, fmt.Errorf("%s are not configured.", strings.Join(errVars, ", "))
	}

	if k := os.Getenv("STATE_STORE_KEY"); k != "" {
		app.StateStoreKey = k
	} else {
		app.StateStoreKey = "tsdakoku:states"
	}

	if k := os.Getenv("ACCESS_TOKEN_STORE_KEY"); k != "" {
		app.AccessTokenStoreKey = k
	} else {
		app.AccessTokenStoreKey = "tsdakoku:access_tokens"
	}

	app.ClientID = clientID
	app.ClientSecret = clientSecret
	app.SlackVerificationToken = slackVerificationToken
	app.TeamSpiritHost = teamSpilitHost
	if err := app.SetupRedis(); err != nil {
		return app, err
	}
	return app, nil
}

func Run() (*App, error) {
	app, err := New()
	if err != nil {
		return app, err
	}
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if !(port > 0) {
		port = 8000
	}
	app.Port = port
	router := app.SetupRouter()
	fmt.Println("Listeninng on 0.0.0.0:" + strconv.Itoa(port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), apachelog.CombinedLog.Wrap(router, os.Stderr)))
	return app, nil
}
