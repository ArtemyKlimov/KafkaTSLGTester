package generator

import (
	"fmt"
	"math/rand"
)

type IPGenerator struct {
	rng *rand.Rand
}

func (g *IPGenerator) Resolve() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		g.rng.Intn(256),
		g.rng.Intn(256),
		g.rng.Intn(256),
		g.rng.Intn(256))
}
