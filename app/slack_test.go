package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/nlopes/slack"
	"golang.org/x/oauth2"
	null "gopkg.in/guregu/null.v3"
	gock "gopkg.in/h2non/gock.v1"
)

func createActionCallbackRequest(callbackID, actionType, token string) *http.Request {
	callback := &slack.AttachmentActionCallback{
		CallbackID: callbackID,
		Actions: []slack.AttachmentAction{{
			Name: actionType,
			SelectedOptions: []slack.AttachmentActionOption{
				{Value: "C1234567"},
			},
		}},
		Token:       token,
		ResponseURL: "https://hooks.slack.test/coolhook",
		User: slack.User{
			ID: "FOO",
		},
	}
	json, _ := json.Marshal(callback)
	data := url.Values{}
	data.Set("payload", string(json))
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/hooks/interactive", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	return req
}

func setupActionCallbackGocks(actionType string, responseText string) {
	if actionType == actionTypeAttend || actionType == actionTypeLeave {
		gock.New("https://teamspirit-1234.cloudforce.test").
			Put("/services/apexrest/Dakoku").
			Reply(200).
			BodyString(responseText)
	} else {
		gock.New("https://teamspirit-1234.cloudforce.test").
			Post("/services/apexrest/Dakoku").
			Reply(200).
			BodyString(responseText)
	}

	gock.New("https://teamspirit-1234.cloudforce.test").
		Get("/services/apexrest/Dakoku").
		Reply(200).
		JSON(map[string]interface{}{
			"timeTable": []map[string]interface{}{{"from": 1, "to": 2, "type": 1}},
			"isHoliday": false,
		})
}

func testGetActionCallbackWithActionType(t *testing.T, actionType string, successMessage string) {
	defer gock.Off()
	app := createMockApp()
	app.CleanRedis()
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/hooks/interactive", strings.NewReader(""))
	ctx := app.createContext(req)
	ctx.UserID = "FOO"
	ctx.setSalesforceAccessToken(&oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: "bar",
		TokenType:    "Bearer",
	})

	gock.New("https://teamspirit-1234.cloudforce.test").
		Get("/services/apexrest/Dakoku").
		Reply(200).
		JSON([]map[string]interface{}{{"message": "Session expired or invalid", "errorCode": "INVALID_SESSION_ID"}})

	gock.InterceptClient(ctx.createTimeTableClient().HTTPClient)

	msg, responseURL, err := ctx.getActionCallback(&slack.AttachmentActionCallback{
		Actions:     []slack.AttachmentAction{{Name: actionType}},
		Token:       app.SlackVerificationToken,
		ResponseURL: "https://hooks.slack.test/coolhook",
		User: slack.User{
			ID: "FOO",
		},
	})

	for _, test := range []Test{
		{true, err == nil},
		{"https://hooks.slack.test/coolhook", responseURL},
		{"TeamSpirit で認証を行って、再度 `/ts` コマンドを実行してください :bow:", msg.Attachments[0].Text},
		{0, strings.Index(msg.Attachments[0].Actions[0].URL, "https://example.com/oauth/salesforce/authenticate/")},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}

	setupActionCallbackGocks(actionType, `"OK"`)
	msg, responseURL, err = ctx.getActionCallback(&slack.AttachmentActionCallback{
		Actions:     []slack.AttachmentAction{{Name: actionType}},
		Token:       app.SlackVerificationToken,
		ResponseURL: "https://hooks.slack.test/coolhook",
		User: slack.User{
			ID: "FOO",
		},
	})

	for _, test := range []Test{
		{true, err == nil},
		{"https://hooks.slack.test/coolhook", responseURL},
		{successMessage, msg.Text},
		{"in_channel", msg.ResponseType},
		{true, msg.ReplaceOriginal},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}

	setupActionCallbackGocks(actionType, "NG")

	msg, responseURL, err = ctx.getActionCallback(&slack.AttachmentActionCallback{
		Actions:     []slack.AttachmentAction{{Name: actionType}},
		Token:       app.SlackVerificationToken,
		ResponseURL: "https://hooks.slack.test/coolhook",
		User: slack.User{
			ID: "FOO",
		},
	})

	for _, test := range []Test{
		{true, err == nil},
		{"https://hooks.slack.test/coolhook", responseURL},
		{"勤務表の更新に失敗しました :warning:", msg.Text},
		{"ephemeral", msg.ResponseType},
		{false, msg.ReplaceOriginal},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
}

func TestGetActionCallback(t *testing.T) {
	testGetActionCallbackWithActionType(t, actionTypeAttend, "出勤しました :office:")
	testGetActionCallbackWithActionType(t, actionTypeLeave, "退勤しました :house:")
	testGetActionCallbackWithActionType(t, actionTypeRest, "休憩を開始しました :coffee:")
	testGetActionCallbackWithActionType(t, actionTypeUnrest, "休憩を終了しました :computer:")
}

func setupTimeTableGocks(items []timeTableItem, isHoliday *bool) {
	gock.New("https://teamspirit-1234.cloudforce.test").
		Get("/services/apexrest/Dakoku").
		Reply(200).
		JSON(map[string]interface{}{
			"timeTable": items,
			"isHoliday": isHoliday,
		})
}

func TestGetSlackMessage(t *testing.T) {
	defer gock.Off()
	app := createMockApp()
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/hooks/slash", bytes.NewBufferString(""))
	ctx := app.createContext(req)
	ctx.UserID = "BAZ"
	msg, err := ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678"})
	for _, test := range []Test{
		{true, err == nil},
		{"TeamSpirit で認証を行って、再度 `/ts` コマンドを実行してください :bow:", msg.Attachments[0].Text},
		{0, strings.Index(msg.Attachments[0].Actions[0].URL, "https://example.com/oauth/salesforce/authenticate/")},
	} {
		test.Compare(t)
	}
	ctx.setSalesforceAccessToken(&oauth2.Token{
		AccessToken:  "foo",
		RefreshToken: "bar",
		TokenType:    "Bearer",
	})
	ctx.TimeTableClient = nil
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678"})
	for _, test := range []Test{
		{true, err == nil},
		{"TeamSpirit で認証を行って、再度 `/ts` コマンドを実行してください :bow:", msg.Attachments[0].Text},
		{0, strings.Index(msg.Attachments[0].Actions[0].URL, "https://example.com/oauth/salesforce/authenticate/")},
	} {
		test.Compare(t)
	}
	setupTimeTableGocks([]timeTableItem{
		{null.IntFrom(10 * 60), null.IntFrom(19 * 60), 1},
	}, nil)
	ctx.TimeTableClient = nil
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678", Text: "channel"})
	for _, test := range []Test{
		{true, err == nil},
		{"Slack で認証を行って、再度 `/ts channel` コマンドを実行してください :bow:", msg.Attachments[0].Text},
		{0, strings.Index(msg.Attachments[0].Actions[0].URL, "https://example.com/oauth/slack/authenticate/T12345678/")},
	} {
		test.Compare(t)
	}
	setupTimeTableGocks([]timeTableItem{
		{null.IntFrom(10 * 60), null.IntFrom(19 * 60), 1},
	}, nil)
	ctx.TimeTableClient = nil
	ctx.setSlackAccessToken("foo")
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678", Text: "channel"})
	for _, test := range []Test{
		{true, err == nil},
		{"打刻時に通知するチャネルを選択して下さい", msg.Attachments[0].Text},
	} {
		test.Compare(t)
	}
	setupTimeTableGocks([]timeTableItem{
		{null.IntFrom(10 * 60), null.IntFrom(19 * 60), 1},
	}, nil)
	ctx.TimeTableClient = nil
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678"})
	for _, test := range []Test{
		{true, err == nil},
		{"既に退勤済です。打刻修正は <https://teamspirit-1234.cloudforce.test|TeamSpirit> で行なってください。", msg.Text},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
	setupTimeTableGocks([]timeTableItem{
		{null.IntFrom(10 * 60), null.IntFromPtr(nil), 1},
		{null.IntFrom(10 * 60), null.IntFromPtr(nil), 21},
	}, &[]bool{false}[0])
	ctx.TimeTableClient = nil
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678"})
	for _, test := range []Test{
		{true, err == nil},
		{"休憩を終了する", msg.Attachments[0].Actions[0].Text},
		{actionTypeUnrest, msg.Attachments[0].Actions[0].Name},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
	setupTimeTableGocks([]timeTableItem{
		{null.IntFrom(10 * 60), null.IntFromPtr(nil), 1},
		{null.IntFrom(10 * 60), null.IntFromPtr(nil), 21},
	}, &[]bool{false}[0])
	ctx.TimeTableClient = nil
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678"})
	for _, test := range []Test{
		{true, err == nil},
		{"休憩を終了する", msg.Attachments[0].Actions[0].Text},
		{actionTypeUnrest, msg.Attachments[0].Actions[0].Name},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
	setupTimeTableGocks([]timeTableItem{
		{null.IntFrom(10 * 60), null.IntFromPtr(nil), 1},
		{null.IntFrom(10 * 60), null.IntFrom(11 * 60), 21},
	}, &[]bool{false}[0])
	ctx.TimeTableClient = nil
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678"})
	for _, test := range []Test{
		{true, err == nil},
		{"休憩を開始する", msg.Attachments[0].Actions[0].Text},
		{actionTypeRest, msg.Attachments[0].Actions[0].Name},
		{"退勤する", msg.Attachments[0].Actions[1].Text},
		{actionTypeLeave, msg.Attachments[0].Actions[1].Name},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
	setupTimeTableGocks([]timeTableItem{
		{null.IntFromPtr(nil), null.IntFromPtr(nil), 1},
		{null.IntFrom(10 * 60), null.IntFrom(11 * 60), 21},
	}, &[]bool{false}[0])
	ctx.TimeTableClient = nil
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678"})
	for _, test := range []Test{
		{true, err == nil},
		{"出勤する", msg.Attachments[0].Actions[0].Text},
		{actionTypeAttend, msg.Attachments[0].Actions[0].Name},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
	setupTimeTableGocks([]timeTableItem{}, &[]bool{true}[0])
	ctx.TimeTableClient = nil
	msg, err = ctx.getSlackMessage(slack.SlashCommand{TeamID: "T12345678"})
	for _, test := range []Test{
		{true, err == nil},
		{"本日は休日です :sunny:", msg.Text},
		{true, gock.IsDone()},
	} {
		test.Compare(t)
	}
}
