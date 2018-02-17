package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	gock "gopkg.in/h2non/gock.v1"
)

func TestSetupRouter(t *testing.T) {
	app := createMockApp()
	router := app.setupRouter()
	paths := []string{}
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, _ := route.GetPathTemplate()
		paths = append(paths, pathTemplate)
		return nil
	})
	Test{[]string{
		"/",
		"/favicon.ico",
		"/success",
		"/oauth/callback",
		"/oauth/authenticate/{state}",
		"/hooks/slash",
		"/hooks/interactive",
	}, paths}.DeepEqual(t)
}

func TestHandleAsset(t *testing.T) {
	app := createMockApp()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/foo", nil)
	app.handleAsset("hoge", res, req)
	for _, test := range []Test{
		{404, res.Code},
		{0, strings.Index(res.Body.String(), "Asset assets/hoge not found")},
		{"text/plain; charset=utf-8", res.Header().Get("Content-Type")},
	} {
		test.Compare(t)
	}
}

func TestHandleIndex(t *testing.T) {
	app := createMockApp()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	app.setupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{200, res.Code},
		{422, strings.Index(res.Body.String(), `<h1 class="cover-heading">ts-dakoku</h1>`)},
		{"text/html; charset=utf-8", res.Header().Get("Content-Type")},
	} {
		test.Compare(t)
	}
}

func TestHandleAuthSuccess(t *testing.T) {
	app := createMockApp()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/success", nil)
	app.setupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{200, res.Code},
		{100, strings.Index(res.Body.String(), "<title>認証完了 - ts-dakoku</title>")},
		{"text/html; charset=utf-8", res.Header().Get("Content-Type")},
	} {
		test.Compare(t)
	}
}

func TestHandleFavicon(t *testing.T) {
	app := createMockApp()
	app.setupRouter()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/favicon.ico", nil)
	app.setupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{200, res.Code},
		{"image/vnd.microsoft.icon", res.Header().Get("Content-Type")},
	} {
		test.Compare(t)
	}
}

func TestHandleAuthenticate(t *testing.T) {
	app := createMockApp()
	app.CleanRedis()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/oauth/authenticate/", nil)
	ctx := app.createContext(req)
	ctx.UserID = "FOO"
	state, _ := ctx.storeUserIDInState()
	req, _ = http.NewRequest(http.MethodGet, "https://example.com/oauth/authenticate/"+state, nil)
	app.setupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{303, res.Code},
		{"https://login.salesforce.com/services/oauth2/authorize?access_type=offline&client_id=SALESFORCE_CLIENT_ID+is+set%21&redirect_uri=https%3A%2F%2Fexample.com%2Foauth%2Fcallback&response_type=code&scope=refresh_token+full&state=" + state, res.Header().Get("Location")},
	} {
		test.Compare(t)
	}
}

func TestHandleAuthenticateNotFound(t *testing.T) {
	app := createMockApp()
	app.CleanRedis()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/oauth/authenticate/foo", nil)
	ctx := app.createContext(req)
	ctx.UserID = "FOO"
	app.setupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{404, res.Code},
	} {
		test.Compare(t)
	}
}

func TestHandleOAuthCallback(t *testing.T) {
	defer gock.Off()
	expiry, _ := time.Parse("2016-01-02T15:04:05Z", "0001-01-01T00:00:00Z")
	gock.New("https://login.salesforce.com").
		Post("/services/oauth2/token").
		Reply(200).
		JSON(oauth2.Token{
			AccessToken:  "foo",
			RefreshToken: "bar",
			TokenType:    "Bearer",
			Expiry:       expiry,
		})
	app := createMockApp()
	app.CleanRedis()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	ctx := app.createContext(req)
	ctx.UserID = "FOO"
	state, _ := ctx.storeUserIDInState()
	token := ctx.getAccessTokenForUser()
	Test{true, token == nil}.Compare(t)
	req, _ = http.NewRequest(http.MethodGet, "https://example.com/oauth/callback?state="+state+"&code=fjkfjk", nil)
	app.setupRouter().ServeHTTP(res, req)
	token = ctx.getAccessTokenForUser()
	for _, test := range []Test{
		{302, res.Code},
		{false, token == nil},
		{"bar", token.RefreshToken},
		{"foo", token.AccessToken},
		{false, token.Expiry.IsZero()},
		{"/success", res.Header().Get("Location")},
	} {
		test.Compare(t)
	}
}

func TestHandleOAuthCallbackError(t *testing.T) {
	defer gock.Off()
	gock.New("https://login.salesforce.com").
		Post("/services/oauth2/token").
		Reply(400).
		BodyString("NG")

	app := createMockApp()
	app.CleanRedis()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	ctx := app.createContext(req)
	ctx.UserID = "FOO"
	state, _ := ctx.storeUserIDInState()
	req, _ = http.NewRequest(http.MethodGet, "https://example.com/oauth/callback?state="+state+"&code=fjkfjk", nil)
	app.setupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{500, res.Code},
	} {
		test.Compare(t)
	}
}

func createSlashCommandRequest(data url.Values) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/hooks/slash", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	return req
}

func TestHandleSlashCommand(t *testing.T) {
	defer gock.Off()
	app := createMockApp()
	app.CleanRedis()
	res := httptest.NewRecorder()
	req := createSlashCommandRequest(url.Values{
		"token": {"hoge"},
	})
	app.setupRouter().ServeHTTP(res, req)
	Test{401, res.Code}.Compare(t)

	res = httptest.NewRecorder()
	req = createSlashCommandRequest(url.Values{
		"token": {app.SlackVerificationToken},
	})
	app.setupRouter().ServeHTTP(res, req)
	time.Sleep(time.Second)
	for _, test := range []Test{
		{200, res.Code},
		{"", res.Body.String()},
		{"text/plain", res.Header().Get("Content-Type")},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
}

func TestHandleActionCallback(t *testing.T) {
	defer gock.Off()
	app := createMockApp()
	app.CleanRedis()

	res := httptest.NewRecorder()
	req := createActionCallbackRequest(actionTypeAttend, "foo")
	app.setupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{401, res.Code},
		{"Invlaid token", strings.Trim(res.Body.String(), "\n\t\r")},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}

	res = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "https://example.com/hooks/interactive", strings.NewReader("payload=[]"))
	app.setupRouter().ServeHTTP(res, req)
	time.Sleep(time.Second)
	for _, test := range []Test{
		{400, res.Code},
		{"unexpected end of JSON input", strings.Trim(res.Body.String(), "\n\t\r")},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}

	gock.New("https://hooks.slack.test").
		Post("/coolhook").
		Reply(200).
		JSON([]map[string]interface{}{{"success": true}})

	gock.New("https://teamspirit-1234.cloudforce.test").
		Get("/services/apexrest/Dakoku").
		Reply(200).
		JSON([]map[string]interface{}{{"from": 1, "to": 2, "type": 1}})

	gock.New("https://teamspirit-1234.cloudforce.test").
		Put("/services/apexrest/Dakoku").
		Reply(200).
		BodyString("OK")

	app.CleanRedis()
	res = httptest.NewRecorder()
	req = createActionCallbackRequest(actionTypeAttend, app.SlackVerificationToken)
	ctx := app.createContext(req)
	ctx.UserID = "FOO"
	ctx.setAccessToken(&oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: "bar",
		TokenType:    "Bearer",
	})

	app.setupRouter().ServeHTTP(res, req)
	time.Sleep(time.Second)
	for _, test := range []Test{
		{200, res.Code},
		{"勤務表を更新中", res.Body.String()},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
}
