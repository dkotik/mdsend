package locks

import (
	"context"
	"time"
)

// var (
//     ErrLockExpired = errors.New("the lock entry has expired")
// )

type Lock interface {
	IsLockedAndLockIfNot(context.Context, []byte) (bool, error)
	Expire(context.Context, time.Time) (int, error)
}
