package pair

import (
	"fmt"
	"math"
	"time"
)

type Pair struct {
	Timestamp time.Time
	Value     float64
}

func (p Pair) String() string {
	return fmt.Sprintf("(%s, %.2f)",
		p.Timestamp.UTC().Format(time.RFC3339), p.Value)
}

func (p Pair) Equals(o Pair, tolerance float64) bool {
	if p.Timestamp != o.Timestamp {
		return false
	}

	if math.Abs(p.Value-o.Value) > tolerance {
		return false
	}

	return true
}
