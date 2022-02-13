package locks

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func runExpiryTestCases(lock Lock, t *testing.T) {
	tcs := []struct {
		Token  []byte
		Locked bool
	}{
		{[]byte("test2"), false},
		{[]byte("test21"), false},
		{[]byte("test22"), false},
		{[]byte("test23"), false},
		{[]byte("test2"), true},
		{[]byte("test22"), true},
		{[]byte("test24"), false},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, c := range tcs {
		expected, err := lock.IsLockedAndLockIfNot(ctx, c.Token)
		if err != nil {
			t.Fatalf("failed test case for key %q", c.Token)
		}
		if expected != c.Locked {
			t.Fatalf("token %q lock did not match the expected value: %+v", c.Token, c.Locked)
		}
	}

	n, err := lock.Expire(ctx, time.Now().Add(expiry))
	if err != nil {
		t.Fatal(fmt.Errorf("could not expire locks: %w", err))
	}
	if n < 5 {
		t.Fatalf("%d entries were expired instead of the expected more than 5", n)
	}
}
