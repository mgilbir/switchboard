package switchboard

import (
	"time"
)

var (
	Now = func() time.Time { return time.Now() }
)

type TimeBin int

func TimeToHourBin(t time.Time) TimeBin {
	return TimeToBin(t, time.Hour)
}

func TimeToBin(t time.Time, d time.Duration) TimeBin {
	return TimeBin(t.Truncate(d).Unix())

}
