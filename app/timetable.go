package app

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type TimeTable struct {
	Items []TimeTableItem `json:"timeTable"`
}

type TimeTableItem struct {
	From int `json:"from"`
	To   int `json:"to"`
	Type int `json:"type"`
}

type TimeTableClient struct {
	AccessToken string
	Endpoint    string
}

func convertTime(time time.Time) int {
	return 0
}

func (tt *TimeTable) IsAttending() bool {
	for _, item := range tt.Items {
		if item.Type == 1 && item.From > 0 {
			return true
		}
	}
	return false
}

func (tt *TimeTable) IsResting() bool {
	for _, item := range tt.Items {
		if (item.Type == 21 || item.Type == 22) && item.To == 0 {
			return true
		}
	}
	return false
}

func (tt *TimeTable) IsLeaving() bool {
	for _, item := range tt.Items {
		if item.Type == 1 && item.To > 0 {
			return true
		}
	}
	return false
}

func (ctx *Context) CreateTimeTableClient() *TimeTableClient {
	return &TimeTableClient{
		AccessToken: ctx.GetAccessTokenForUser(),
		Endpoint:    "https://" + ctx.TeamSpiritHost + "/services/apexrest/Dakoku",
	}
}

func (client *TimeTableClient) doRequest(method string, data io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, client.Endpoint, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+client.AccessToken)
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	httpClient := &http.Client{}
	res, err := httpClient.Do(req)
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
	var items []TimeTableItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	return &TimeTable{Items: items}, nil
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
