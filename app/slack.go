package app

import (
	"time"

	"github.com/nlopes/slack"
)

const (
	actionTypeAttend           = "attend"
	actionTypeRest             = "rest"
	actionTypeUnrest           = "unrest"
	actionTypeLeave            = "leave"
	actionTypeSelectChannel    = "select-channel"
	actionTypeUnselectChannel  = "unselect-channel"
	callbackIDChannelSelect    = "slack_channel_select_button"
	callbackIDAttendanceButton = "attendance_button"
)

func (ctx *Context) getActionCallback(data *slack.AttachmentActionCallback) (*slack.Msg, string, error) {
	ctx.UserID = data.User.ID
	client := ctx.createTimeTableClient()
	timeTable, err := client.GetTimeTable()
	if err != nil {
		state := State{
			TeamID:      data.Team.ID,
			UserID:      ctx.UserID,
			ResponseURL: data.ResponseURL,
		}
		err, msg := ctx.getLoginSlackMessage(state)
		return err, data.ResponseURL, msg
	}

	text := ""
	now := time.Now()
	attendance := -1
	switch data.Actions[0].Name {
	case actionTypeLeave:
		{
			attendance = 0
			text = "退勤しました :house:"
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
			text = "出勤しました :office:"
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

func (ctx *Context) getLoginSlackMessage(state State) (*slack.Msg, error) {
	stateKey, err := ctx.storeState(state)
	if err != nil {
		return nil, err
	}
	return &slack.Msg{
		Attachments: []slack.Attachment{
			slack.Attachment{
				Text:       "TeamSpirit で認証を行って、再度 `/ts` コマンドを実行してください :bow:",
				CallbackID: callbackIDAttendanceButton,
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "authenticate",
						Value: "authenticate",
						Text:  "認証する",
						Style: "primary",
						Type:  "button",
						URL:   ctx.getSalesforceAuthenticateURL(stateKey),
					},
				},
			},
		},
	}, nil
}

func (ctx *Context) getAuthenticateSlackMessage(state State) (*slack.Msg, error) {
	stateKey, err := ctx.storeState(state)
	if err != nil {
		return nil, err
	}
	return &slack.Msg{
		Attachments: []slack.Attachment{
			slack.Attachment{
				Text:       "Slack で認証を行って、再度 `/ts channel` コマンドを実行してください :bow:",
				CallbackID: "slack_authentication_button",
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "slack-authenticate",
						Value: "slack-authenticate",
						Text:  "認証する",
						Style: "primary",
						Type:  "button",
						URL:   ctx.getSlackAuthenticateURL(state.TeamID, stateKey),
					},
				},
			},
		},
	}, nil
}

func (ctx *Context) getChannelSelectSlackMessage() (*slack.Msg, error) {
	return &slack.Msg{
		Attachments: []slack.Attachment{
			slack.Attachment{
				Text:       "打刻時に通知するチャネルを選択して下さい",
				CallbackID: callbackIDChannelSelect,
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:       actionTypeSelectChannel,
						Value:      actionTypeSelectChannel,
						Text:       "チャネルを選択",
						Type:       "select",
						DataSource: "channels",
					},
					slack.AttachmentAction{
						Name:  actionTypeUnrest,
						Value: actionTypeUnrest,
						Text:  "通知を止める",
						Style: "danger",
						Type:  "button",
					},
				},
			},
		},
	}, nil
}

func (ctx *Context) getSlackMessage(command slack.SlashCommand) (*slack.Msg, error) {
	text := command.Text
	state := State{
		TeamID:      command.TeamID,
		UserID:      command.UserID,
		ResponseURL: command.ResponseURL,
	}
	client := ctx.createTimeTableClient()
	if client.HTTPClient == nil || text == "login" {
		return ctx.getLoginSlackMessage(state)
	}
	timeTable, err := client.GetTimeTable()
	if err != nil {
		return ctx.getLoginSlackMessage(state)
	}
	if text == "channel" {
		if ctx.getSlackAccessTokenForUser() == "" {
			return ctx.getAuthenticateSlackMessage(state)
		}
		return ctx.getChannelSelectSlackMessage()
	}
	if timeTable.IsLeaving() {
		return &slack.Msg{
			Text: "既に退勤済です。打刻修正は <https://" + ctx.TeamSpiritHost + "|TeamSpirit> で行なってください。",
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
					CallbackID: callbackIDAttendanceButton,
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
					CallbackID: callbackIDAttendanceButton,
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
							Text:  "退勤する",
							Style: "danger",
							Type:  "button",
							Confirm: &slack.ConfirmationField{
								Text:        "退勤しますか？",
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
				CallbackID: callbackIDAttendanceButton,
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  actionTypeAttend,
						Value: actionTypeAttend,
						Text:  "出勤する",
						Style: "primary",
						Type:  "button",
					},
				},
			},
		},
	}, nil

}
