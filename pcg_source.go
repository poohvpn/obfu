package obfu

import (
	"math/rand"

	"github.com/MichaelTJones/pcg"
)

type pcgSource struct {
	*pcg.PCG64
}

var _ rand.Source64 = &pcgSource{}

func newPcgSource(seed int64) *pcgSource {
	src := &pcgSource{
		PCG64: pcg.NewPCG64(),
	}
	src.Seed(seed)
	return src
}

func (p *pcgSource) Int63() int64 {
	return int64(p.PCG64.Random() >> 1)
}

func (p *pcgSource) Uint64() uint64 {
	return p.PCG64.Random()
}

func (p *pcgSource) Seed(seed int64) {
	p.PCG64.Seed(uint64(seed), 14468578777317145133, 7178091608504981567, uint64(^seed))
}
