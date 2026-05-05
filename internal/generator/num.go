package generator

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type NumGenerator struct {
	min, max int
	rng      *rand.Rand
}

func NewNumGenerator(raw string) *NumGenerator {
	parts := strings.SplitN(raw, "to", 2)
	min, _ := strconv.Atoi(parts[0])
	max := min
	if len(parts) == 2 {
		max, _ = strconv.Atoi(parts[1])
	}
	if max < min {
		min, max = max, min
	}
	return &NumGenerator{min: min, max: max, rng: newRand()}
}

func (g *NumGenerator) Resolve() string {
	n := g.min
	if g.max > g.min {
		n = g.min + g.rng.Intn(g.max-g.min+1)
	}
	return fmt.Sprintf("%d", n)
}
