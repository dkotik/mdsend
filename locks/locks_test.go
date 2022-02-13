package locks

import (
	"context"
	"crypto/sha256"
	"flag"
	"hash"
	"path/filepath"
	"testing"
	"time"
)

const expiry = time.Second

var integration = flag.Bool("integration", false, "run integration tests")

func TestAnonymized(t *testing.T) {
	t.Parallel()

	lock, closer, err := NewBBoltLock(
		filepath.Join(t.TempDir(), "test.bbolt"), "bucket", expiry)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closer(); err != nil {
			t.Fatal(err)
		}
	}()
	runTestCases(&Anonymized{
		Lock:  lock,
		noise: []byte("testnoise"),
		hashFactory: func() hash.Hash {
			return sha256.New()
		},
	}, t)
}

func TestBBolt(t *testing.T) {
	t.Parallel()

	// TODO: t.Fatal("test Expire call")

	lock, closer, err := NewBBoltLock(
		filepath.Join(t.TempDir(), "test.bbolt"), "bucket", expiry)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closer(); err != nil {
			t.Fatal(err)
		}
	}()
	runTestCases(lock, t)
	runExpiryTestCases(lock, t)
}

func runTestCases(lock Lock, t *testing.T) {
	tcs := []struct {
		Token  []byte
		Locked bool
	}{
		{[]byte("test"), false},
		{[]byte("test1"), false},
		{[]byte("test2"), false},
		{[]byte("test3"), false},
		{[]byte("test"), true},
		{[]byte("test2"), true},
		{[]byte("test4"), false},
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

	if !*integration {
		return
	}

	time.Sleep(expiry)

	for _, c := range tcs {
		if !c.Locked {
			continue
		}
		expected, err := lock.IsLockedAndLockIfNot(ctx, c.Token)
		if err != nil {
			t.Fatalf("failed test case for key %q", c.Token)
		}
		if expected != false {
			t.Fatalf("token %q should be unlocked but it is not", c.Token)
		}
	}
}
