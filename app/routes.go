package app

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *App) SetupRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", app.HandleIndex).Methods("GET")
	router.HandleFunc("/favicon.ico", app.HandleFavicon).Methods("GET")
	router.HandleFunc("/oauth/callback", app.HandleOAuthCallback).Methods("GET")
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

func (app *App) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	token, err := app.GetAccessToken(code, state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	context := app.CreateContext(r)
	context.SetAccessToken(token)
	http.Redirect(w, r, "/", http.StatusFound)
}
