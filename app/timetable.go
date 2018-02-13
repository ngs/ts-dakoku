package app

type TimeTable struct {
	Items []TimeTableItem `json:"timeTable"`
}

type TimeTableItem struct {
	From int `json:"from"`
	To   int `json:"to"`
	Type int `json:"type"`
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
