package generator

import (
	"crypto/rand"
	"encoding/hex"
)

// HexGenerator produces N random lowercase hex characters.
type HexGenerator struct {
	n int
}

func NewHexGenerator(n int) *HexGenerator {
	if n <= 0 {
		n = 1
	}
	return &HexGenerator{n: n}
}

func (g *HexGenerator) Resolve() string {
	bytesNeeded := (g.n + 1) / 2
	b := make([]byte, bytesNeeded)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:g.n]
}
