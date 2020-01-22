package util

import (
	"strconv"
	"time"
)

func TimeFmt(tp interface{}, sample string, tz int) string {
	offset := int64(time.Duration(tz) * time.Hour)
	var t int64
	switch tp.(type) {
	case uint64:
		t = int64(tp.(uint64))
	case string:
		i64, err := strconv.ParseInt(tp.(string), 10, 64)
		if err != nil {
			return ""
		}
		t = i64
	case int64:
		t = tp.(int64)
	}
	if len(sample) == 0 {
		sample = "2006-01-02"
	}
	tm := time.Unix(t, offset).UTC()
	return tm.Format(sample)
}
