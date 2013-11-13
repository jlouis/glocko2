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

func volK(f func(float64) float64, a float64, tau float64) float64 {
	k := 0.0
	c := a - k * math.Sqrt(tau * tau)
	for ; f(c) < 0.0; k += 1.0 {
		c = a - k * math.Sqrt(tau * tau)
	}
	
	return c
}

func sign(x float64) float64 {
	if x < 0 {
		return -1.0
	} else if x > 0 {
		return 1.0
	} else {
		return 0.0
	}
}

func computeVolatility(sigma float64, phi float64, v float64, delta float64, tau float64) float64 {
	a := math.Log(sigma * sigma)
	phi2 := phi * phi
	f := func (x float64) float64 {
		ex := math.Exp(x)
		d2 := delta * delta
		a2 := phi2 + v + ex
		p2 := (x - a) / (tau * tau)
		p1 := (ex * (d2 - phi2 - v - ex)) / (2 * a2 * a2)
		return (p1 - p2)
	}
	
	var b float64
	if delta * delta > phi * phi {
		b = math.Log(delta * delta - phi * phi - v)
	} else {
		b  = volK(f, a, tau)
	}
	
	fa := f(a)
	fb := f(b)
	
	var c, fc, d, fd float64
	for i := 100; ; i-- {
		if math.Abs(b - a) <= ε {
			return math.Exp(a / 2)
		} else {
			c = (a + b) * 0.5
			fc = f(c)
			d = c + (c - a) * (sign(fa - fb) * fc) / math.Sqrt(fc * fc - fa*fb)
			fd = f(d)
			
			if sign(fd) != sign(fc) {
				a = c
				b = d
				fa = fc
				fb = fd
			} else if sign(fd) != sign(fa) {
				b = d
				fb = fd
			} else {
				a = d
				fa = fd
			}
		}
	}
	
	panic("Exceeded iterations")
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
