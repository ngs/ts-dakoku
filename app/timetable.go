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

type timeTable struct {
	Items     []timeTableItem `json:"timeTable"`
	IsHoliday *bool           `json:"isHoliday,omitempty"`
}

type timeTableItem struct {
	From null.Int `json:"from,omitempty"`
	To   null.Int `json:"to,omitempty"`
	Type int      `json:"type"`
}

type timeTableError struct {
	Message string `json:"message"`
	Code    string `json:"errorCode"`
}

type timeTableClient struct {
	HTTPClient *http.Client
	Endpoint   string
}

func parseTimeTable(body []byte) (*timeTable, error) {
	var errors []timeTableError
	if err := json.Unmarshal(body, &errors); err == nil && len(errors) > 0 && errors[0].Code != "" {
		return nil, fmt.Errorf("Error: %+v (%+v)", errors[0].Message, errors[0].Code)
	}
	var timeTable timeTable
	if err := json.Unmarshal(body, &timeTable); err != nil {
		return nil, err
	}
	return &timeTable, nil
}

func convertTime(time time.Time) null.Int {
	hour, min, _ := time.Clock()
	return null.IntFrom(int64(hour*60 + min))
}

func (item *timeTableItem) IsAttendance() bool {
	return item.Type == 1
}

func (item *timeTableItem) IsRest() bool {
	return item.Type == 21 || item.Type == 22
}

func (tt *timeTable) IsAttending() bool {
	for _, item := range tt.Items {
		if item.IsAttendance() && item.From.Valid {
			return true
		}
	}
	return false
}

func (tt *timeTable) IsResting() bool {
	for _, item := range tt.Items {
		if item.IsRest() && !item.To.Valid {
			return true
		}
	}
	return false
}

func (tt *timeTable) IsLeaving() bool {
	for _, item := range tt.Items {
		if item.IsAttendance() && item.To.Valid {
			return true
		}
	}
	return false
}

func (tt *timeTable) Attend(time time.Time) bool {
	items := tt.Items
	for i, item := range items {
		if item.IsAttendance() {
			items[i].From = convertTime(time)
			tt.Items = items
			return true
		}
	}
	tt.Items = append(tt.Items, timeTableItem{
		From: convertTime(time),
		Type: 1,
	})
	return true
}

func (tt *timeTable) Rest(time time.Time) bool {
	tt.Items = append(tt.Items, timeTableItem{
		From: convertTime(time),
		Type: 21,
	})
	return true
}

func (tt *timeTable) Unrest(time time.Time) bool {
	items := tt.Items
	for i, item := range items {
		if item.IsRest() && !item.To.Valid {
			items[i].To = convertTime(time)
			tt.Items = items
			return true
		}
	}
	tt.Items = append(tt.Items, timeTableItem{
		To:   convertTime(time),
		Type: 21,
	})
	return true

}

func (tt *timeTable) Leave(time time.Time) bool {
	items := tt.Items
	for i, item := range items {
		if item.Type == 1 {
			items[i].To = convertTime(time)
			tt.Items = items
			return true
		}
	}
	tt.Items = append(tt.Items, timeTableItem{
		To:   convertTime(time),
		Type: 1,
	})
	return true
}

func (ctx *Context) createTimeTableClient() *timeTableClient {
	if ctx.TimeTableClient != nil {
		return ctx.TimeTableClient
	}
	ctx.TimeTableClient = &timeTableClient{
		HTTPClient: ctx.getSalesforceOAuth2Client(),
		Endpoint:   "https://" + ctx.TeamSpiritHost + "/services/apexrest/Dakoku",
	}
	return ctx.TimeTableClient
}

func (client *timeTableClient) doRequest(method string, data io.Reader) ([]byte, error) {
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
	body, err := ioutil.ReadAll(res.Body)
	return body, err
}

func (client *timeTableClient) GetTimeTable() (*timeTable, error) {
	body, err := client.doRequest(http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	return parseTimeTable(body)
}

func (client *timeTableClient) UpdateTimeTable(timeTable *timeTable) (bool, error) {
	timeTable.IsHoliday = nil
	b, err := json.Marshal(timeTable)
	if err != nil {
		return false, err
	}
	body, err := client.doRequest(http.MethodPost, bytes.NewBuffer(b))
	fmt.Printf("%v %v %v\n", string(body), err, string(body) == `"OK"`)
	if err != nil {
		return false, err
	}
	return string(body) == `"OK"`, nil
}

func (client *timeTableClient) SetAttendance(attendance bool) (bool, error) {
	data := map[string]bool{"attendance": attendance}
	b, err := json.Marshal(data)
	if err != nil {
		return false, err
	}
	body, err := client.doRequest(http.MethodPut, bytes.NewBuffer(b))
	if err != nil {
		return false, err
	}
	return string(body) == `"OK"`, nil
}
