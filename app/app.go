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
	Port                          int
	ClientSecret                  string
	ClientID                      string
	SlashCommandVerificationToken string
	RedisConn                     redis.Conn
}

func New() (*App, error) {
	app := &App{}
	clientSecret := os.Getenv("SALESFORCE_CLIENT_SECRET")
	clientID := os.Getenv("SALESFORCE_CLIENT_ID")
	slashCommandVerificationToken := os.Getenv("SLASH_COMMAND_VERIFICATION_TOKEN")
	var errVars = []string{}
	if clientSecret == "" {
		errVars = append(errVars, "SALESFORCE_CLIENT_SECRET")
	}
	if clientID == "" {
		errVars = append(errVars, "SALESFORCE_CLIENT_ID")
	}
	if slashCommandVerificationToken == "" {
		errVars = append(errVars, "SLASH_COMMAND_VERIFICATION_TOKEN")
	}
	if len(errVars) > 0 {
		return app, fmt.Errorf("%s are not configured.", strings.Join(errVars, ", "))
	}
	app.ClientID = clientID
	app.ClientSecret = clientSecret
	app.SlashCommandVerificationToken = slashCommandVerificationToken
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
