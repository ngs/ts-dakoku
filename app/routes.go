package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
)

func (app *App) SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", app.HandleIndex).Methods("GET")
	router.HandleFunc("/favicon.ico", app.HandleFavicon).Methods("GET")
	router.HandleFunc("/oauth/callback", app.HandleOAuthCallback).Methods("GET")
	router.HandleFunc("/oauth/authenticate/{state}", app.HandleAuthenticate).Methods("GET")
	router.HandleFunc("/hooks/slash", app.HandleSlashCommand).Methods("POST")
	return router
}

func (app *App) HandleIndex(w http.ResponseWriter, r *http.Request) {
	app.handleAsset("index.html", w, r)
}

func (app *App) HandleFavicon(w http.ResponseWriter, r *http.Request) {
	app.handleAsset("favicon.ico", w, r)
}

func (app *App) handleAsset(filename string, w http.ResponseWriter, r *http.Request) {
	data, err := Asset("assets/" + filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write(data)
	}
}

func (app *App) HandleAuthenticate(w http.ResponseWriter, r *http.Request) {
	app.ReconnectRedisIfNeeeded()
	vars := mux.Vars(r)
	state := vars["state"]
	ctx := app.CreateContext(r)
	if userID := ctx.GetUserIDForState(state); userID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	config := ctx.GetOAuth2Config()
	url := config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (app *App) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	app.ReconnectRedisIfNeeeded()
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	ctx := app.CreateContext(r)
	token, err := ctx.GetAccessToken(code, state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.UserID = ctx.GetUserIDForState(state)
	ctx.SetAccessToken(token)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *App) HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	app.ReconnectRedisIfNeeeded()
	s, err := slack.SlashCommandParse(r)
	fmt.Printf("Command: %v Error:%v\n", s, err)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !s.ValidateToken(app.SlashCommandVerificationToken) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx := app.CreateContext(r)
	ctx.UserID = s.UserID

	params := &slack.Msg{}

	if ctx.GetAccessTokenForUser() == "" || s.Text == "login" {
		state, err := ctx.StoreUserIDInState()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		params.Attachments = []slack.Attachment{
			slack.Attachment{
				Text: "TeamSpirit で認証を行ってください",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "authenticate",
						Text:  "認証する",
						Style: "primary",
						Type:  "button",
						URL:   ctx.GetAuthenticateURL(state),
					},
				},
			},
		}
	}

	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
