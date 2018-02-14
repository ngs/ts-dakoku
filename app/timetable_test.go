package app

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	null "gopkg.in/guregu/null.v3"
)

type Test struct {
	expected interface{}
	actual   interface{}
}

func getMockTime() time.Time {
	loc := time.FixedZone("Asia/Tokyo", 9*60*60)
	return time.Date(2018, time.September, 1, 11, 12, 22, 0, loc)
}

func (test Test) Compare(t *testing.T) {
	if test.expected != test.actual {
		t.Errorf(`Expected "%v" but got "%v"`, test.expected, test.actual)
	}
}

func (test Test) DeepEqual(t *testing.T) {
	if !reflect.DeepEqual(test.expected, test.actual) {
		t.Errorf(`Expected "%v" but got "%v"`, test.expected, test.actual)
	}
}

func TestMarshalTimeTable(t *testing.T) {
	b, _ := json.Marshal(TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 3},
			{From: null.IntFrom(4), Type: 1},
		},
	})
	Test{`{"timeTable":[{"from":1,"to":2,"type":3},{"from":4,"to":null,"type":1}]}`, string(b)}.Compare(t)
}

func TestParseTimeTable(t *testing.T) {
	timeTable, err := ParseTimeTable([]byte(`[{"from":600, "to": null, "type": 1}, {"from":780, "to": 840, "type": 21}]`))
	for _, test := range []Test{
		{nil, err},
		{2, len(timeTable.Items)},
		{int64(600), timeTable.Items[0].From.ValueOrZero()},
		{false, timeTable.Items[0].To.Valid},
		{1, timeTable.Items[0].Type},
		{int64(780), timeTable.Items[1].From.ValueOrZero()},
		{int64(840), timeTable.Items[1].To.ValueOrZero()},
		{21, timeTable.Items[1].Type},
	} {
		test.Compare(t)
	}
}

func TestParseTimeTableError(t *testing.T) {
	timeTable, err := ParseTimeTable([]byte(`[{"message":"Session expired or invalid","errorCode":"INVALID_SESSION_ID"}]`))
	for _, test := range []Test{
		{"Error: Session expired or invalid (INVALID_SESSION_ID)", err.Error()},
		{true, timeTable == nil},
	} {
		test.Compare(t)
	}
}

func TestIsAttending(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 1},
		},
	}
	Test{true, tt.IsAttending()}.Compare(t)
	tt.Items[2].Type = 22
	Test{false, tt.IsAttending()}.Compare(t)
}

func TestIsResting(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 1},
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), Type: 21},
		},
	}
	Test{true, tt.IsResting()}.Compare(t)
	tt.Items[2].To = null.IntFrom(3)
	Test{false, tt.IsResting()}.Compare(t)
}

func TestIsLeaving(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), Type: 1},
		},
	}
	Test{false, tt.IsLeaving()}.Compare(t)
	tt.Items[2].To = null.IntFrom(1)
	Test{true, tt.IsLeaving()}.Compare(t)
}

func TestAttend(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
		},
	}
	res := tt.Attend(getMockTime())
	for _, test := range []Test{
		{3, len(tt.Items)},
		{res, true},
		{int64(672), tt.Items[2].From.ValueOrZero()},
		{false, tt.Items[2].To.Valid},
		{1, tt.Items[2].Type},
	} {
		test.Compare(t)
	}
}

func TestAttend2(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{To: null.IntFrom(1140), Type: 1},
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
		},
	}
	res := tt.Attend(getMockTime())
	for _, test := range []Test{
		{3, len(tt.Items)},
		{res, true},
		{int64(672), tt.Items[0].From.ValueOrZero()},
		{int64(1140), tt.Items[0].To.ValueOrZero()},
		{1, tt.Items[0].Type},
	} {
		test.Compare(t)
	}
}

func TestRest(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), Type: 21},
		},
	}
	res := tt.Rest(getMockTime())
	for _, test := range []Test{
		{3, len(tt.Items)},
		{res, true},
		{int64(672), tt.Items[2].From.ValueOrZero()},
		{false, tt.Items[2].To.Valid},
		{21, tt.Items[2].Type},
	} {
		test.Compare(t)
	}
}

func TestUnrest(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), Type: 21},
		},
	}
	res := tt.Unrest(getMockTime())
	for _, test := range []Test{
		{2, len(tt.Items)},
		{res, true},
		{true, tt.Items[1].To.Valid},
		{int64(672), tt.Items[1].To.ValueOrZero()},
		{21, tt.Items[1].Type},
	} {
		test.Compare(t)
	}
}

func TestLeave(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), Type: 21},
		},
	}
	res := tt.Leave(getMockTime())
	for _, test := range []Test{
		{3, len(tt.Items)},
		{res, true},
		{true, tt.Items[2].To.Valid},
		{int64(672), tt.Items[2].To.ValueOrZero()},
		{false, tt.Items[2].From.Valid},
		{1, tt.Items[2].Type},
	} {
		test.Compare(t)
	}
}

func TestLeave2(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: null.IntFrom(1), To: null.IntFrom(2), Type: 21},
			{From: null.IntFrom(1), Type: 21},
			{From: null.IntFrom(600), Type: 1},
		},
	}
	res := tt.Leave(getMockTime())
	for _, test := range []Test{
		{3, len(tt.Items)},
		{res, true},
		{true, tt.Items[2].To.Valid},
		{int64(672), tt.Items[2].To.ValueOrZero()},
		{int64(600), tt.Items[2].From.ValueOrZero()},
		{1, tt.Items[2].Type},
	} {
		test.Compare(t)
	}
}

func TestConvertTime(t *testing.T) {
	result := convertTime(getMockTime())
	Test{int64(672), result.ValueOrZero()}.Compare(t)
}
