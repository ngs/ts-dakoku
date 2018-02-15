package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/guregu/null.v3"
)

type TimeTable struct {
	Items []TimeTableItem `json:"timeTable"`
}

type TimeTableItem struct {
	From null.Int `json:"from,omitempty"`
	To   null.Int `json:"to,omitempty"`
	Type int      `json:"type"`
}

type TimeTableError struct {
	Message string `json:"message"`
	Code    string `json:"errorCode"`
}

type TimeTableClient struct {
	HTTPClient *http.Client
	Endpoint   string
}

func ParseTimeTable(body []byte) (*TimeTable, error) {
	var errors []TimeTableError
	if err := json.Unmarshal(body, &errors); err == nil && len(errors) > 0 && errors[0].Code != "" {
		return nil, fmt.Errorf("Error: %+v (%+v)", errors[0].Message, errors[0].Code)
	}
	var items []TimeTableItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	return &TimeTable{Items: items}, nil
}

func convertTime(time time.Time) null.Int {
	hour, min, _ := time.Clock()
	return null.IntFrom(int64(hour*60 + min))
}

func (item *TimeTableItem) IsAttendance() bool {
	return item.Type == 1
}

func (item *TimeTableItem) IsRest() bool {
	return item.Type == 21 || item.Type == 22
}

func (tt *TimeTable) IsAttending() bool {
	for _, item := range tt.Items {
		if item.IsAttendance() && item.From.Valid {
			return true
		}
	}
	return false
}

func (tt *TimeTable) IsResting() bool {
	for _, item := range tt.Items {
		if item.IsRest() && !item.To.Valid {
			return true
		}
	}
	return false
}

func (tt *TimeTable) IsLeaving() bool {
	for _, item := range tt.Items {
		if item.IsAttendance() && item.To.Valid {
			return true
		}
	}
	return false
}

func (tt *TimeTable) Attend(time time.Time) bool {
	items := tt.Items
	for i, item := range items {
		if item.IsAttendance() {
			items[i].From = convertTime(time)
			tt.Items = items
			return true
		}
	}
	tt.Items = append(tt.Items, TimeTableItem{
		From: convertTime(time),
		Type: 1,
	})
	return true
}

func (tt *TimeTable) Rest(time time.Time) bool {
	tt.Items = append(tt.Items, TimeTableItem{
		From: convertTime(time),
		Type: 21,
	})
	return true
}

func (tt *TimeTable) Unrest(time time.Time) bool {
	items := tt.Items
	for i, item := range items {
		if item.IsRest() && !item.To.Valid {
			items[i].To = convertTime(time)
			tt.Items = items
			return true
		}
	}
	tt.Items = append(tt.Items, TimeTableItem{
		To:   convertTime(time),
		Type: 21,
	})
	return true

}

func (tt *TimeTable) Leave(time time.Time) bool {
	items := tt.Items
	for i, item := range items {
		if item.Type == 1 {
			items[i].To = convertTime(time)
			tt.Items = items
			return true
		}
	}
	tt.Items = append(tt.Items, TimeTableItem{
		To:   convertTime(time),
		Type: 1,
	})
	return true
}

func (ctx *Context) CreateTimeTableClient() *TimeTableClient {
	if ctx.TimeTableClient != nil {
		return ctx.TimeTableClient
	}
	ctx.TimeTableClient = &TimeTableClient{
		HTTPClient: ctx.GetOAuth2Client(),
		Endpoint:   "https://" + ctx.TeamSpiritHost + "/services/apexrest/Dakoku",
	}
	return ctx.TimeTableClient
}

func (client *TimeTableClient) doRequest(method string, data io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, client.Endpoint, data)
	if err != nil {
		return nil, err
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(res.Body)
}

func (client *TimeTableClient) GetTimeTable() (*TimeTable, error) {
	body, err := client.doRequest("GET", nil)
	if err != nil {
		return nil, err
	}
	return ParseTimeTable(body)
}

func (client *TimeTableClient) UpdateTimeTable(timeTable *TimeTable) (bool, error) {
	b, err := json.Marshal(timeTable)
	if err != nil {
		return false, err
	}
	body, err := client.doRequest("POST", bytes.NewBuffer(b))
	if err != nil {
		return false, err
	}
	return string(body) == "OK", nil
}

func (client *TimeTableClient) SetAttendance(attendance bool) (bool, error) {
	data := map[string]bool{"attendance": attendance}
	b, err := json.Marshal(data)
	if err != nil {
		return false, err
	}
	body, err := client.doRequest("PUT", bytes.NewBuffer(b))
	if err != nil {
		return false, err
	}
	return string(body) == "OK", nil
}
