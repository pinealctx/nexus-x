package nxutil

import (
	"strconv"
	"time"
)

// UnixNowString returns the current Unix timestamp as a string.
func UnixNowString() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// UnixMsNow returns the current time as Unix milliseconds.
func UnixMsNow() int64 {
	return time.Now().UnixMilli()
}
