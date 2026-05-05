package generator

import (
	"math/rand"
	"strings"
)

type OneOfGenerator struct {
	choices []string
	rng     *rand.Rand
}

func NewOneOfGenerator(raw string) *OneOfGenerator {
	return &OneOfGenerator{
		choices: strings.Split(raw, ","),
		rng:     newRand(),
	}
}

func (g *OneOfGenerator) Resolve() string {
	return g.choices[g.rng.Intn(len(g.choices))]
}
