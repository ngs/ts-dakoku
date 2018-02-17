package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
)

func (app *App) setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", app.handleIndex).Methods(http.MethodGet)
	router.HandleFunc("/favicon.ico", app.handleFavicon).Methods(http.MethodGet)
	router.HandleFunc("/success", app.handleAuthSuccess).Methods(http.MethodGet)
	router.HandleFunc("/oauth/callback", app.handleOAuthCallback).Methods(http.MethodGet)
	router.HandleFunc("/oauth/authenticate/{state}", app.handleAuthenticate).Methods(http.MethodGet)
	router.HandleFunc("/hooks/slash", app.handleSlashCommand).Methods(http.MethodPost)
	router.HandleFunc("/hooks/interactive", app.handleActionCallback).Methods(http.MethodPost)
	return router
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	app.handleAsset("index.html", w, r)
}

func (app *App) handleAuthSuccess(w http.ResponseWriter, r *http.Request) {
	app.handleAsset("success.html", w, r)
}

func (app *App) handleFavicon(w http.ResponseWriter, r *http.Request) {
	app.handleAsset("favicon.ico", w, r)
}

func (app *App) handleAsset(filename string, w http.ResponseWriter, r *http.Request) {
	data, err := Asset("assets/" + filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else {
		w.Write(data)
	}
}

func (app *App) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	app.reconnectRedisIfNeeeded()
	vars := mux.Vars(r)
	state := vars["state"]
	ctx := app.createContext(r)
	if userID := ctx.getUserIDForState(state); userID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	config := ctx.getOAuth2Config()
	config.Scopes = []string{"refresh_token", "full"}
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (app *App) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	app.reconnectRedisIfNeeeded()
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	ctx := app.createContext(r)
	token, err := ctx.getAccessToken(code, state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.UserID = ctx.getUserIDForState(state)
	ctx.setAccessToken(token)
	ctx.deleteState(state)
	http.Redirect(w, r, "/success", http.StatusFound)
}

func (app *App) handleSlashCommand(w http.ResponseWriter, r *http.Request) {
	app.reconnectRedisIfNeeeded()
	s, err := slack.SlashCommandParse(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !s.ValidateToken(app.SlackVerificationToken) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx := app.createContext(r)
	ctx.UserID = s.UserID

	go func() {
		params, _ := ctx.getSlackMessage(s.Text)
		b, _ := json.Marshal(params)
		http.Post(s.ResponseURL, "application/json", bytes.NewBuffer(b))
	}()

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(""))
}

func (app *App) handleActionCallback(w http.ResponseWriter, r *http.Request) {
	app.reconnectRedisIfNeeeded()
	ctx := app.createContext(r)
	r.ParseForm()
	payload := r.PostForm.Get("payload")

	var data slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if data.Token != ctx.SlackVerificationToken {
		http.Error(w, "Invlaid token", http.StatusUnauthorized)
		return
	}
	go func() {
		params, responseURL, err := ctx.getActionCallback(&data)
		if err != nil && params == nil && responseURL != "" {
			http.Post(responseURL, "text/plain", bytes.NewBufferString(err.Error()))
			return
		} else if err != nil {
			fmt.Printf("Handle Action Callback Error: %+v\n", err.Error())
		}
		b, _ := json.Marshal(params)
		http.Post(responseURL, "application/json", bytes.NewBuffer(b))
	}()

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("勤務表を更新中"))
}
