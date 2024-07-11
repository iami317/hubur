package hubur

import "time"

func TimeStampNano() int64 {
	return time.Now().UnixNano()
}

func TimeStampMs() int64 {
	return TimeStampNano() / int64(time.Millisecond)
}

func TimeStampSecond() int64 {
	return time.Now().Unix()
}

func DatetimePretty() string {
	return time.Now().Format("2006_01_02-15_04_05")
}
