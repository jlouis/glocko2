package glocko2

import (
	"math"
)

const (
	scaling = 173.7178 // Scaling factor for Glicko2
	ε       = 0.000001
	tau     = 0.5
)

type Player struct {
	R     float64 // Player ranking
	Rd    float64 // Ranking deviation
	Sigma float64 // Volatility
}

type Opponent struct {
	Rj  float64 // Opponent ranking
	Rdj float64 // Opponent ranking deviation
	Sj  float64 // Score— 0.0, 1.0 or 0.5 for Loss, Win, Draw respectively
}

type opp struct {
	muj   float64
	phij  float64
	gphij float64
	emmp  float64
	sj    float64
}

func scale(r float64, rd float64) (mu float64, phi float64) {
	mu = (r - 1500.0) / scaling
	phi = (rd / scaling)
	return mu, phi
}

func g(phi float64) float64 {
	return (1 / math.Sqrt(1+3*phi*phi/(math.Pi*math.Pi)))
}

func e(mu float64, muj float64, phij float64) float64 {
	return (1 / (1 + math.Exp(-g(phij)*(mu-muj))))
}

func scaleOpponents(mu float64, os []Opponent) (res []opp) {
	res = make([]opp, len(os))
	for i, o := range os {
		muj, phij := scale(o.Rj, o.Rdj)
		res[i] = opp{muj, phij, g(phij), e(mu, muj, phij), o.Sj}
	}

	return res
}

func updateRating(sopp []opp) float64 {
	s := 0.0
	for _, o := range sopp {
		s += o.gphij * o.gphij * o.emmp * (1 - o.emmp)
	}

	return 1 / s
}

func computeDelta(v float64, sopp []opp) float64 {
	s := 0.0
	for _, o := range sopp {
		s += o.gphij * (o.sj - o.emmp)
	}

	return v * s
}

func computeVolatility(sigma float64, phi float64, v float64, delta float64, tau float64) float64 {
	return 0.0
}

func phiStar(sigmap float64, phi float64) float64 {
	return math.Sqrt(phi*phi + sigmap*sigmap)
}

func newRating(phis float64, mu float64, v float64, sopp []opp) (float64, float64) {
	phip := 1 / math.Sqrt(
			(1 / (phis * phis)) + (1 / v))
	s := 0.0
	for _, o := range sopp {
		s += o.gphij * (o.sj - o.emmp)
	}
	mup := mu + (phip*phip) * s
	return mup, phip
}

func unscale(mup float64, phip float64) (float64, float64) {
	rp := scaling * mup + 1500.0
	rdp := scaling * phip
	return rp, rdp
}