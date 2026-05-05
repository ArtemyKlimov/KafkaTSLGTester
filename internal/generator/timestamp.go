package generator

import (
	"fmt"
	"time"
)

// TimestampGenerator formats the current UTC time.
// precision: 0=seconds, 3=milliseconds, 6=microseconds, 9=nanoseconds.
type TimestampGenerator struct {
	precision int
}

func (g *TimestampGenerator) Resolve() string {
	t := time.Now().UTC()
	base := t.Format("2006-01-02T15:04:05")
	switch g.precision {
	case 3:
		return fmt.Sprintf("%s.%03dZ", base, t.Nanosecond()/1_000_000)
	case 6:
		return fmt.Sprintf("%s.%06dZ", base, t.Nanosecond()/1_000)
	case 9:
		return fmt.Sprintf("%s.%09dZ", base, t.Nanosecond())
	default:
		return base + "Z"
	}
}
