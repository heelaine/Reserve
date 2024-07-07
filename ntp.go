package main

import (
	"github.com/beevik/ntp"
	"time"
)

func GetNTPOffset() (*time.Duration, error) {
	q, err := ntp.Query("ntp.aliyun.com")
	return &q.ClockOffset, err
}
