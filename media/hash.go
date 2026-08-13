package media

import (
	"bytes"
	"io"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/cespare/xxhash/v2"
)

func DeterministicHashStringOf(data []byte) string {
	h := xxhash.New()
	_, _ = io.Copy(h, bytes.NewReader(data))
	return base58.Encode(h.Sum(nil))
}
