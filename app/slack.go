package app

import "github.com/nlopes/slack"

const (
	ActionTypeAttend = "attend"
	ActionTypeRest   = "rest"
	ActionTypeUnrest = "unrest"
	ActionTypeLeave  = "leave"
)

func (ctx *Context) GetSlackMessage(text string) (*slack.Msg, error) {
	client := ctx.CreateTimeTableClient()
	if client.AccessToken == "" || text == "login" {
		state, err := ctx.StoreUserIDInState()
		if err != nil {
			return nil, err
		}
		return &slack.Msg{
			Attachments: []slack.Attachment{
				slack.Attachment{
					Text:       "TeamSpirit で認証を行ってください",
					CallbackID: "authentication_button",
					Actions: []slack.AttachmentAction{
						slack.AttachmentAction{
							Name:  "authenticate",
							Value: "authenticate",
							Text:  "認証する",
							Style: "primary",
							Type:  "button",
							URL:   ctx.GetAuthenticateURL(state),
						},
					},
				},
			},
		}, nil
	}
	timeTable, err := client.GetTimeTable()
	if err != nil {
		return nil, err
	}
	if timeTable.IsLeaving() {
		return &slack.Msg{
			Text: "既に退社済です。打刻修正は TeamSpirit で行なってください。",
		}, nil
	}
	if timeTable.IsResting() {
		return &slack.Msg{
			Attachments: []slack.Attachment{
				slack.Attachment{
					CallbackID: "attendance_button",
					Actions: []slack.AttachmentAction{
						slack.AttachmentAction{
							Name:  ActionTypeUnrest,
							Value: ActionTypeUnrest,
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
							Name:  ActionTypeRest,
							Value: ActionTypeRest,
							Text:  "休憩を開始する",
							Style: "default",
							Type:  "button",
						},
						slack.AttachmentAction{
							Name:  ActionTypeLeave,
							Value: ActionTypeLeave,
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
						Name:  ActionTypeAttend,
						Value: ActionTypeAttend,
						Text:  "出社する",
						Style: "primary",
						Type:  "button",
					},
				},
			},
		},
	}, nil

}
