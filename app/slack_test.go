package app

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/nlopes/slack"
)

func createActionCallbackRequest(actionName string, token string) *http.Request {
	callback := &slack.AttachmentActionCallback{
		Actions: []slack.AttachmentAction{{Name: actionName}},
		Token:   token,
	}
	json, _ := json.Marshal(callback)
	data := url.Values{}
	data.Set("payload", string(json))
	req, _ := http.NewRequest("POST", "https://example.com/hooks/interactive", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	return req
}

func TestGetActionCallback(t *testing.T) {
	app := createMockApp()

	ctx := app.CreateContext(createActionCallbackRequest(ActionTypeLeave, "hoge"))
	msg, err := ctx.GetActionCallback()
	for _, test := range []Test{
		{"Invalid token", err.Error()},
		{true, msg == nil},
	} {
		test.Compare(t)
	}

	ctx = app.CreateContext(createActionCallbackRequest(ActionTypeLeave, app.SlackVerificationToken))
	msg, err = ctx.GetActionCallback()
	for _, test := range []Test{
		{true, err == nil},
		{"TeamSpirit で認証を行って、再度 `/ts` コマンドを実行してください :bow:", msg.Attachments[0].Text},
		{0, strings.Index(msg.Attachments[0].Actions[0].URL, "https://example.com/oauth/authenticate/")},
	} {
		test.Compare(t)
	}
}

func TestGetLoginSlackMessage(t *testing.T) {
	t.Skip("TODO")
}

func TestGetSlackMessage(t *testing.T) {
	t.Skip("TODO")
}
