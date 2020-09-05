package gym

import (
	"fmt"
	"time"
)

type Utilization struct {
	Timestamp time.Time
	People    uint64
	Capacity  uint64
}

func (u *Utilization) String() string {
	return fmt.Sprintf("(%s, %d, %d)",
		u.Timestamp.Format(time.RFC3339), u.People, u.Capacity)
}

func (u *Utilization) Percent() (float64, bool) {
	if u.Capacity == 0 {
		return 0.0, false
	}

	return 100.0 * float64(u.People) / float64(u.Capacity), true
}

func (u *Utilization) Equal(o *Utilization) bool {
	if !u.Timestamp.Equal(o.Timestamp) {
		return false
	}

	if u.People != o.People {
		return false
	}

	if u.Capacity != o.Capacity {
		return false
	}

	return true
}
