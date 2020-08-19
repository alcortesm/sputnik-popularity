package pair

import (
	"fmt"
	"time"
)

type Pair struct {
	Timestamp time.Time
	Value     float64
}

func (p Pair) String() string {
	return fmt.Sprintf("(%s, %f)",
		p.Timestamp.UTC().Format(time.RFC3339), p.Value)
}
