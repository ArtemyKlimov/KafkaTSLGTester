package generator

import (
	"math/rand"
	"strings"
)

const alphanumChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomTextGenerator produces N random alphanumeric characters.
type RandomTextGenerator struct {
	n   int
	rng *rand.Rand
}

func NewRandomTextGenerator(n int) *RandomTextGenerator {
	if n <= 0 {
		n = 1
	}
	return &RandomTextGenerator{n: n, rng: newRand()}
}

func (g *RandomTextGenerator) Resolve() string {
	b := make([]byte, g.n)
	for i := range b {
		b[i] = alphanumChars[g.rng.Intn(len(alphanumChars))]
	}
	return string(b)
}

// RandomWordGenerator concatenates random words from the word list until
// the result is at least N characters, then trims to exactly N.
type RandomWordGenerator struct {
	n     int
	words []string
	rng   *rand.Rand
}

func NewRandomWordGenerator(n int, words []string) *RandomWordGenerator {
	if n <= 0 {
		n = 1
	}
	if len(words) == 0 {
		words = []string{"word"}
	}
	return &RandomWordGenerator{n: n, words: words, rng: newRand()}
}

func (g *RandomWordGenerator) Resolve() string {
	var sb strings.Builder
	for sb.Len() < g.n {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(g.words[g.rng.Intn(len(g.words))])
	}
	return sb.String()[:g.n]
}
