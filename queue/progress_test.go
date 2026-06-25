package queue_test

import (
	"testing"
)

func TestProgressTracker(t *testing.T) {
	if testing.Short() {
		t.Skip("progress tracker reads and writes many records")
	}

}
