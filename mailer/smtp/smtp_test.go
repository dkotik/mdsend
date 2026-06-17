package smtp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	sqliteQ "github.com/dkotik/mdsend/queue/sqlite"
	"zombiezen.com/go/sqlite"
)

var getLiveConfigOrSkip func(t *testing.T) Configuration

func TestMain(m *testing.M) {
	conn, err := sqlite.OpenConn(":memory:")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = conn.Close(); err != nil {
			panic(err)
		}
	}()
	q, err := sqliteQ.New(conn, "")
	if err != nil {
		panic(err)
	}

	liveConfig, err := Configuration{
		Queue: q,
	}.withDefaults()
	if err != nil {
		getLiveConfigOrSkip = func(t *testing.T) Configuration {
			t.Skip("live test config error: " + err.Error())
			return liveConfig
		}
	} else {
		liveTestLockFile := filepath.Join("testdata", "live_test.lock")
		_, err := os.Stat(liveTestLockFile)
		if err == nil {
			getLiveConfigOrSkip = func(t *testing.T) Configuration {
				t.Skip("live test lock file exists: remove it to allow live tests")
				return liveConfig
			}
		} else {
			if !os.IsNotExist(err) {
				panic(fmt.Errorf("failed to check live test lock file: %w", err))
			}
			if err = os.WriteFile(liveTestLockFile, []byte("delete this file to fire off live test"), 0600); err != nil {
				panic(fmt.Errorf("failed to create live test lock file: %w", err))
			}
			getLiveConfigOrSkip = func(t *testing.T) Configuration {
				return liveConfig
			}
		}
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}
