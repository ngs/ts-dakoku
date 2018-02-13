package app

import "github.com/nlopes/slack"

func (ctx *Context) GetSlackMessage(text string) (*slack.Msg, error) {
	if ctx.GetAccessTokenForUser() == "" || text == "login" {
		state, err := ctx.StoreUserIDInState()
		if err != nil {
			return nil, err
		}
		return &slack.Msg{
			Attachments: []slack.Attachment{
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
			},
		}, nil
	}
	return &slack.Msg{
		Attachments: []slack.Attachment{
			slack.Attachment{
				Actions: []slack.AttachmentAction{
					slack.AttachmentAction{
						Name:  "rest",
						Text:  "休憩を開始する",
						Style: "default",
						Type:  "button",
					},
					slack.AttachmentAction{
						Name:  "leave",
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
