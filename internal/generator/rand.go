package generator

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
)

func newRand() *rand.Rand {
	var b [8]byte
	_, _ = crand.Read(b[:])
	seed := int64(binary.LittleEndian.Uint64(b[:]))
	return rand.New(rand.NewSource(seed)) //nolint:gosec
}
