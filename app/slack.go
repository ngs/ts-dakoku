package app

import (
	"time"

	"github.com/nlopes/slack"
)

const (
	actionTypeAttend = "attend"
	actionTypeRest   = "rest"
	actionTypeUnrest = "unrest"
	actionTypeLeave  = "leave"
)

func (ctx *Context) getActionCallback(data *slack.AttachmentActionCallback) (*slack.Msg, string, error) {
	ctx.UserID = data.User.ID
	client := ctx.createTimeTableClient()
	timeTable, err := client.GetTimeTable()
	if err != nil {
		err, msg := ctx.getLoginSlackMessage()
		return err, data.ResponseURL, msg
	}

	text := ""
	now := time.Now()
	attendance := -1
	switch data.Actions[0].Name {
	case actionTypeLeave:
		{
			attendance = 0
			text = "退社しました :house:"
		}
	case actionTypeRest:
		{
			timeTable.Rest(now)
			text = "休憩を開始しました :coffee:"
		}
	case actionTypeUnrest:
		{
			timeTable.Unrest(now)
			text = "休憩を終了しました :computer:"
		}
	case actionTypeAttend:
		{
			attendance = 1
			text = "出社しました :office:"
		}
	}

	params := &slack.Msg{
		ResponseType:    "in_channel",
		ReplaceOriginal: true,
		Text:            text,
	}

	var ok bool
	if attendance != -1 {
		ok, err = client.SetAttendance(attendance == 1)
	} else {
		ok, err = client.UpdateTimeTable(timeTable)
	}
	if !ok || err != nil {
		params.ResponseType = "ephemeral"
		params.ReplaceOriginal = false
		params.Text = "勤務表の更新に失敗しました :warning:"
	}

	return params, data.ResponseURL, nil
}

func (ctx *Context) getLoginSlackMessage() (*slack.Msg, error) {
	state, err := ctx.storeUserIDInState()
	if err != nil {
		return nil, err
	}
	return &slack.Msg{
		Attachments: []slack.Attachment{
			slack.Attachment{
				Text:       "TeamSpirit で認証を行って、再度 `/ts` コマンドを実行してください :bow:",
				CallbackID: "authentication_button",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "authenticate",
						Value: "authenticate",
						Text:  "認証する",
						Style: "primary",
						Type:  "button",
						URL:   ctx.getAuthenticateURL(state),
					},
				},
			},
		},
	}, nil
}

func (ctx *Context) getSlackMessage(text string) (*slack.Msg, error) {
	client := ctx.createTimeTableClient()
	if client.HTTPClient == nil || text == "login" {
		return ctx.getLoginSlackMessage()
	}
	timeTable, err := client.GetTimeTable()
	if err != nil {
		return ctx.getLoginSlackMessage()
	}
	if timeTable.IsLeaving() {
		return &slack.Msg{
			Text: "既に退社済です。打刻修正は <https://" + ctx.TeamSpiritHost + "|TeamSpirit> で行なってください。",
		}, nil
	}
	if timeTable.IsHoliday != nil && *timeTable.IsHoliday == true {
		return &slack.Msg{
			Text: "本日は休日です :sunny:",
		}, nil
	}
	if timeTable.IsResting() {
		return &slack.Msg{
			Attachments: []slack.Attachment{
				slack.Attachment{
					CallbackID: "attendance_button",
					Actions: []slack.AttachmentAction{
						slack.AttachmentAction{
							Name:  actionTypeUnrest,
							Value: actionTypeUnrest,
							Text:  "休憩を終了する",
							Style: "default",
							Type:  "button",
						},
					},
				},
			},
		}, nil
	}
	if timeTable.IsAttending() {
		return &slack.Msg{
			Attachments: []slack.Attachment{
				slack.Attachment{
					CallbackID: "attendance_button",
					Actions: []slack.AttachmentAction{
						slack.AttachmentAction{
							Name:  actionTypeRest,
							Value: actionTypeRest,
							Text:  "休憩を開始する",
							Style: "default",
							Type:  "button",
						},
						slack.AttachmentAction{
							Name:  actionTypeLeave,
							Value: actionTypeLeave,
							Text:  "退社する",
							Style: "danger",
							Type:  "button",
							Confirm: &slack.ConfirmationField{
								Text:        "退社しますか？",
								OkText:      "はい",
								DismissText: "いいえ",
							},
						},
					},
				},
			},
		}, nil
	}
	return &slack.Msg{
		Attachments: []slack.Attachment{
			slack.Attachment{
				CallbackID: "attendance_button",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  actionTypeAttend,
						Value: actionTypeAttend,
						Text:  "出社する",
						Style: "primary",
						Type:  "button",
					},
				},
			},
		},
	}, nil

}
