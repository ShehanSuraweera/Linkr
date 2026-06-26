package domain

import "time"

type ClickEvent struct {
	LinkID    int64
	At        time.Time
	IPHash    string
	UserAgent string // parsed in pipeline, never persisted raw
	Referer   string // parsed in pipeline, never persisted raw
}

type DailyClickStat struct {
	Day   string `db:"day"   json:"day"`
	Count int64  `db:"count" json:"count"`
}

type DeviceStat struct {
	Device string `db:"device" json:"device"`
	Count  int64  `db:"count"  json:"count"`
}

type BrowserStat struct {
	Browser string `db:"browser" json:"browser"`
	Count   int64  `db:"count"   json:"count"`
}

type RefererStat struct {
	Domain string `db:"domain" json:"domain"`
	Count  int64  `db:"count"  json:"count"`
}

type LinkStats struct {
	TotalClicks int64            `json:"total_clicks"`
	Daily       []DailyClickStat `json:"daily"`
	Devices     []DeviceStat     `json:"devices"`
	Browsers    []BrowserStat    `json:"browsers"`
	Referers    []RefererStat    `json:"referers"`
}

// LinkClickStat is a per-link summary used in the analytics overview top-links chart.
type LinkClickStat struct {
	ShortCode   string `db:"short_code"   json:"short_code"`
	TotalClicks int64  `db:"total_clicks" json:"total_clicks"`
}

// OverviewStats aggregates analytics across all links owned by a user.
type OverviewStats struct {
	TotalLinks  int64            `json:"total_links"`
	ActiveLinks int64            `json:"active_links"`
	TotalClicks int64            `json:"total_clicks"`
	Daily       []DailyClickStat `json:"daily"`
	Devices     []DeviceStat     `json:"devices"`
	Browsers    []BrowserStat    `json:"browsers"`
	Referers    []RefererStat    `json:"referers"`
	TopLinks    []LinkClickStat  `json:"top_links"`
}
