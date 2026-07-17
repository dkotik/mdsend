package queue

import "errors"

var skip = errors.New("message skipped when Queuing")

func NewSkipSentinelError() error {
	return skip
}

func IsSkipSentinelError(err error) bool {
	return errors.Is(err, skip)
}
