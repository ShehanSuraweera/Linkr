package domain

import "time"

type ClickEvent struct {
	LinkID    int64
	At        time.Time
	IPHash    string
	UserAgent string
	Referer   string
}

type DailyClickStat struct {
	Day   string `db:"day"   json:"day"`
	Count int64  `db:"count" json:"count"`
}

type LinkStats struct {
	TotalClicks int64            `json:"total_clicks"`
	Daily       []DailyClickStat `json:"daily"`
}
