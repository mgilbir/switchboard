package switchboard

import (
	"time"
)

var (
	Now = func() time.Time { return time.Now() }
)
