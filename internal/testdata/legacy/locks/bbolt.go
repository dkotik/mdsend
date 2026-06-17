package locks

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

func NewBBoltLock(file, namespace string, d time.Duration) (*BBoltLock, func() error, error) {
	if d < time.Second {
		return nil, nil, errors.New("BBoltLock does not track durations of smaller than one second")
	}

	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, nil, err
	}

	bucket := []byte(namespace)

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not create bucket %s: %w", namespace, err)
	}

	return &BBoltLock{
		db:       db,
		bucket:   bucket,
		duration: d,
	}, db.Close, nil
}

type BBoltLock struct {
	db       *bolt.DB
	bucket   []byte
	duration time.Duration
}

func (b *BBoltLock) IsLockedAndLockIfNot(ctx context.Context, token []byte) (ok bool, err error) {
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if v := bucket.Get(token); v != nil {
			ts := binary.BigEndian.Uint32(v)
			// spew.Dump(ts, uint32(time.Now().Unix()))
			if time.Unix(int64(ts), 0).After(time.Now()) {
				ok = true
				return nil
			}
			return bucket.Delete(token) // delete expired entry
		}
		tb := make([]byte, 4)
		binary.BigEndian.PutUint32(tb, uint32(time.Now().Add(b.duration).Unix()))
		return bucket.Put(token, tb)
	})
	return
}

func (b *BBoltLock) Expire(ctx context.Context, cutoff time.Time) (n int, err error) {
	buckets := [][]byte{b.bucket}

	for _, bucket := range buckets {
		if err = b.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(bucket)
			return bucket.ForEach(func(k, v []byte) error {
				ts := binary.BigEndian.Uint32(v)
				if time.Unix(int64(ts), 0).Before(cutoff) {
					n++
					if err := bucket.Delete(k); err != nil {
						return err
					}
				}
				return nil
			})
		}); err != nil {
			return 0, err
		}
	}
	return
}
