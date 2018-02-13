package app

import (
	"encoding/json"
	"reflect"
	"testing"
)

type Test struct {
	expected interface{}
	actual   interface{}
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
			{From: 1, To: 2, Type: 3},
		},
	})
	Test{string(b), `{"timeTable":[{"from":1,"to":2,"type":3}]}`}.Compare(t)
}

func TestUnmershalTimeTableItem(t *testing.T) {
	var items []TimeTableItem
	json.Unmarshal([]byte(`[{"from":600, "to": 1140, "type": 1}, {"from":780, "to": 840, "type": 21}]`), &items)
	for _, test := range []Test{
		{2, len(items)},
		{600, items[0].From},
		{1140, items[0].To},
		{1, items[0].Type},
		{780, items[1].From},
		{840, items[1].To},
		{21, items[1].Type},
	} {
		test.Compare(t)
	}
}

func TestIsAttending(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: 1, To: 2, Type: 21},
			{From: 1, To: 2, Type: 21},
			{From: 1, To: 2, Type: 1},
		},
	}
	Test{true, tt.IsAttending()}.Compare(t)
	tt.Items[2].Type = 22
	Test{false, tt.IsAttending()}.Compare(t)
}

func TestIsResting(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: 1, To: 2, Type: 1},
			{From: 1, To: 2, Type: 21},
			{From: 1, Type: 21},
		},
	}
	Test{true, tt.IsResting()}.Compare(t)
	tt.Items[2].To = 3
	Test{false, tt.IsResting()}.Compare(t)
}

func TestIsLeaving(t *testing.T) {
	tt := TimeTable{
		Items: []TimeTableItem{
			{From: 1, To: 2, Type: 21},
			{From: 1, To: 2, Type: 21},
			{From: 1, Type: 1},
		},
	}
	Test{false, tt.IsLeaving()}.Compare(t)
	tt.Items[2].To = 1
	Test{true, tt.IsLeaving()}.Compare(t)
}
