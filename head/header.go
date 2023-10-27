package head

import (
	"time"
	"strings"
)

type Header struct {
	Key 	string
	Value 	string
}

func GMT_unix_time(unix_time int64) string {
	return strings.Replace(time.Unix(unix_time, 0).Format(time.RFC1123), "UTC", "GMT", 1)
}