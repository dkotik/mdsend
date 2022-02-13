package locks

import (
	"context"
	"hash"
)

// Anonymized lock scrambles the keys before passing them unto a wrapped lock.
type Anonymized struct {
	Lock
	noise       []byte
	hashFactory func() hash.Hash
}

func (a *Anonymized) IsLockedAndLockIfNot(ctx context.Context, token []byte) (ok bool, err error) {
	h := a.hashFactory()
	_, err = h.Write(a.noise)
	if err != nil {
		return false, err
	}
	_, err = h.Write(token)
	if err != nil {
		return false, err
	}
	return a.Lock.IsLockedAndLockIfNot(ctx, h.Sum(nil))
}
