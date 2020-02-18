// Package glocko2 implements Glicko2 calculations in Go
//
// This package implements the Glicko2 chess rating system in Go. It can be
// used for any 1v1 game in order to rank players on a strength scale.
package glocko2

import (
	"math"
)

const (
	scaling = 173.7178 // Scaling factor for Glicko2
	ε       = 0.000001
)

// Player represents a player in the ranking
type Player struct {
	// Identification of the player. Unique.
	Id string
	// Player name
	Name string
	// Current rank, R, of the player
	R float64
	// Ranking deviation, the systems confidence in R
	Rd float64
	// Volatility of the ranking. How surprising it is the ranking system.
	Sigma float64
	// True if the player is currently actively playing games
	Active bool
}

// Opponent represents an opponent for the player
type Opponent struct {
	// Player index into an []Player array
	Idx int
	// Match score
	Sj float64
}

type opp struct {
	muj   float64
	phij  float64
	gphij float64
	emmp  float64
	sj    float64
}

// Scale transforms the external rating values into the internal rating values
func Scale(r float64, rd float64) (mu float64, phi float64) {
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

func scaleOpponents(mu float64, os []Opponent, players []Player) (res []opp) {
	res = make([]opp, len(os))
	for i, o := range os {
		muj, phij := Scale(players[o.Idx].R, players[o.Idx].Rd)
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
	c := a - k*math.Sqrt(tau*tau)
	i := 0
	for ; f(c) < 0.0; k += 1.0 {
		c = a - k*math.Sqrt(tau*tau)
		i++
		if i > 10000 {
			panic("volK exceeded")
		}
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

// Computing volatility requires us to find a root of a function. We
// use a numerically stable root finding method here. In the Glicko2
// systems, there have been some variants of this historically, because
// some root finding methods aren't numerically stable.
func computeVolatility(sigma float64, phi float64, v float64, delta float64, tau float64) float64 {
	a := math.Log(sigma * sigma)
	phi2 := phi * phi
	f := func(x float64) float64 {
		ex := math.Exp(x)
		d2 := delta * delta
		a2 := phi2 + v + ex
		p2 := (x - a) / (tau * tau)
		p1 := (ex * (d2 - phi2 - v - ex)) / (2 * a2 * a2)
		return (p1 - p2)
	}

	var b float64
	if delta*delta > phi*phi+v {
		b = math.Log(delta*delta - phi*phi - v)
	} else {
		b = volK(f, a, tau)
	}

	fa := f(a)
	fb := f(b)

	var c, fc, d, fd float64
	for i := 100; i > 0; i-- {
		if math.Abs(b-a) <= ε {
			return math.Exp(a / 2)
		}

		c = (a + b) * 0.5
		fc = f(c)
		d = c + (c-a)*(sign(fa-fb)*fc)/math.Sqrt(fc*fc-fa*fb)
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

	panic("Exceeded iterations")
}

// PhiStar computes rating deviations for inactive players
func PhiStar(sigmap float64, phi float64) float64 {
	return math.Sqrt(phi*phi + sigmap*sigmap)
}

func newRating(phis float64, mu float64, v float64, sopp []opp) (float64, float64) {
	phip := 1 / math.Sqrt(
		(1/(phis*phis))+(1/v))
	s := 0.0
	for _, o := range sopp {
		s += o.gphij * (o.sj - o.emmp)
	}
	mup := mu + (phip*phip)*s
	return mup, phip
}

// Unscale reverts the transformation done by the Scale function
func Unscale(mup float64, phip float64) (float64, float64) {
	rp := scaling*mup + 1500.0
	rdp := scaling * phip
	return rp, rdp
}

// Rank computes the new rank of player p
//
// Requires a database of players and a list of matches in opponents
func (p *Player) Rank(opponents []Opponent, players []Player, tau float64) (float64, float64, float64) {

	mu, phi := Scale(p.R, p.Rd)
	sopps := scaleOpponents(mu, opponents, players)
	v := updateRating(sopps)
	delta := computeDelta(v, sopps)

	sigmap := computeVolatility(p.Sigma, phi, v, delta, tau)
	phistar := PhiStar(sigmap, phi)
	mup, phip := newRating(phistar, mu, v, sopps)
	r1, rd1 := Unscale(mup, phip)

	return r1, rd1, sigmap
}
