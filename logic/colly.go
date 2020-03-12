package logic

import "time"

type CollyParser struct {
	d time.Duration // 抓取间隔，避免过快
}
