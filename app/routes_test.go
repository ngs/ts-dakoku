package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	gock "gopkg.in/h2non/gock.v1"
)

func TestSetupRouter(t *testing.T) {
	app := createMockApp()
	router := app.SetupRouter()
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

func TestHandleIndex(t *testing.T) {
	app := createMockApp()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	app.SetupRouter().ServeHTTP(res, req)
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
	app.SetupRouter().ServeHTTP(res, req)
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
	app.SetupRouter()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/favicon.ico", nil)
	app.SetupRouter().ServeHTTP(res, req)
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
	ctx := app.CreateContext(req)
	ctx.UserID = "FOO"
	state, _ := ctx.StoreUserIDInState()
	req, _ = http.NewRequest(http.MethodGet, "https://example.com/oauth/authenticate/"+state, nil)
	app.SetupRouter().ServeHTTP(res, req)
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
	ctx := app.CreateContext(req)
	ctx.UserID = "FOO"
	app.SetupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{404, res.Code},
	} {
		test.Compare(t)
	}
}

func TestHandleOAuthCallback(t *testing.T) {
	defer gock.Off()
	gock.New("https://login.salesforce.com").
		Post("/services/oauth2/token").
		Reply(200).
		JSON(oauth2.Token{
			AccessToken:  "foo",
			RefreshToken: "bar",
			TokenType:    "Bearer",
		})

	app := createMockApp()
	app.CleanRedis()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	ctx := app.CreateContext(req)
	ctx.UserID = "FOO"
	state, _ := ctx.StoreUserIDInState()
	req, _ = http.NewRequest(http.MethodGet, "https://example.com/oauth/callback?state="+state+"&code=fjkfjk", nil)
	app.SetupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{302, res.Code},
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
	ctx := app.CreateContext(req)
	ctx.UserID = "FOO"
	state, _ := ctx.StoreUserIDInState()
	req, _ = http.NewRequest(http.MethodGet, "https://example.com/oauth/callback?state="+state+"&code=fjkfjk", nil)
	app.SetupRouter().ServeHTTP(res, req)
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
	app := createMockApp()
	app.CleanRedis()
	res := httptest.NewRecorder()
	req := createSlashCommandRequest(url.Values{
		"token": {"hoge"},
	})
	app.SetupRouter().ServeHTTP(res, req)
	Test{401, res.Code}.Compare(t)

	res = httptest.NewRecorder()
	req = createSlashCommandRequest(url.Values{
		"token": {app.SlackVerificationToken},
	})
	app.SetupRouter().ServeHTTP(res, req)
	for _, test := range []Test{
		{200, res.Code},
		{48, strings.Index(res.Body.String(), `"callback_id":"authentication_button"`)},
	} {
		test.Compare(t)
	}
}

func TestHandleActionCallback(t *testing.T) {
	defer gock.Off()
	app := createMockApp()
	app.CleanRedis()
	res := httptest.NewRecorder()
	req := createActionCallbackRequest(ActionTypeAttend, "foo")
	app.SetupRouter().ServeHTTP(res, req)
	Test{401, res.Code}.Compare(t)

	// gock.New("https://teamspirit-1234.cloudforce.test").
	// 	Get("/services/apexrest/Dakoku").
	// 	Reply(200).
	// 	JSON([]map[string]interface{}{{"from": 1, "to": 2, "type": 1}})
	// app.CleanRedis()
	// res = httptest.NewRecorder()
	// req = createActionCallbackRequest(ActionTypeAttend, app.SlackVerificationToken)
	// ctx.UserID = "FOO"
	// ctx.SetAccessToken(&oauth2.Token{
	// 	AccessToken:  "foo",
	// 	RefreshToken: "bar",
	// 	TokenType:    "Bearer",
	// })
	//
	// app.SetupRouter().ServeHTTP(res, req)
	// for _, test := range []Test{
	// 	{200, res.Code},
	// 	{48, strings.Index(res.Body.String(), `"callback_id":"authentication_button"`)},
	// } {
	// 	test.Compare(t)
	// }
}
