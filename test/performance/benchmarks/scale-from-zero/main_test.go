package main

import (
	"fmt"
	"testing"
	"time"
)

func TestBadSLA(t *testing.T) {
	// String returns a string representing the duration in the form "72h3m0.5s".
	// Leading zero units are omitted. As a special case, durations less than one
	// second format use a smaller unit (milli-, micro-, or nanoseconds) to ensure
	// that the leading digit is non-zero. The zero duration formats as 0s.

	t1 := time.Duration(15 * time.Millisecond)
	t2 := time.Duration(27*time.Millisecond + 500*time.Microsecond)

	e := fmt.Errorf("SLA 1 failed. P95 latency is not in %v-%v time range: %s\n", 0, t1, t2)

	fmt.Printf("%s", e.Error())
	e = fmt.Errorf("SLA 1 failed. P95 latency is not in %d-%d time range: %s", 0, t1, t2)
	fmt.Printf("%s", e.Error())
}
