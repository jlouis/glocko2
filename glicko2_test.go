package glocko2

import (
	"math"
	"testing"
)

var (
	epsilon = 0.00000001
)

func TestScale(t *testing.T) {
	const r, mu, rd, phi = 1500.0, 0.0, 200.0, 1.1512924985234674
	x, y := Scale(r, rd)

	if x != mu {
		t.Errorf("scale(%v, ⋯) = %v, ⋯, want %v", r, x, mu)
	}

	if y != phi {
		t.Errorf("scale(⋯, %v) = ⋯, %v, want %v", rd, y, phi)
	}
}

func TestScaleOpponents(t *testing.T) {
	mu := 0.0
	os := []Opponent{
		Opponent{1400, 30, 1},
		Opponent{1550, 100, 0},
		Opponent{1700, 300, 0}}

	scaled := scaleOpponents(mu, os)

	if scaled[0].muj != -0.5756462492617337 ||
		scaled[0].phij != 0.1726938747785201 ||
		scaled[0].gphij != 0.9954980064506083 ||
		scaled[0].emmp != 0.6394677305521533 ||
		scaled[0].sj != 1.0 {
		t.Errorf("scaled[0] = %v", scaled[0])
	}

	if scaled[1].muj != 0.28782312463086684 ||
		scaled[1].phij != 0.5756462492617337 ||
		scaled[1].gphij != 0.9531489778689763 ||
		scaled[1].emmp != 0.4318423561076679 ||
		scaled[1].sj != 0.0 {
		t.Errorf("scaled[1] = %v", scaled[1])
	}

	if scaled[2].muj != 1.1512924985234674 ||
		scaled[2].phij != 1.726938747785201 ||
		scaled[2].gphij != 0.7242354780877526 ||
		scaled[2].emmp != 0.30284072909521925 ||
		scaled[2].sj != 0.0 {
		t.Errorf("scaled[1] = %v", scaled[2])
	}
}

func TestUpdateRating(t *testing.T) {
	const expect, mu = 1.7789770897239976, 0.0
	os := []Opponent{
		Opponent{1400, 30, 1},
		Opponent{1550, 100, 0},
		Opponent{1700, 300, 0}}

	scaled := scaleOpponents(mu, os)

	v := updateRating(scaled)

	if v != expect {
		t.Errorf("updateRating(%v) = %v, want %v", scaled, v, expect)
	}
}

func TestComputeDelta(t *testing.T) {
	const expect, mu, v = -0.4839332609836549, 0.0, 1.7789770897239976
	os := []Opponent{
		Opponent{1400, 30, 1},
		Opponent{1550, 100, 0},
		Opponent{1700, 300, 0}}

	scaled := scaleOpponents(mu, os)

	delta := computeDelta(v, scaled)

	if delta != expect {
		t.Errorf("computeDelta(%v, %v) = %v, want %v", v, scaled, delta, expect)
	}
}

func TestComputeVolatility(t *testing.T) {
	const expect, sigma, phi, v, delta, tau = 0.059995984286488495, 0.06, 1.1512924985234674, 1.7789770897239976, -0.4839332609836549, 0.5

	sigmap := computeVolatility(sigma, phi, v, delta, tau)

	if math.Abs(sigmap-expect) > epsilon {
		t.Errorf("computeVolatility(⋯) = %v, want %v", sigmap, expect)
	}
}

func TestPhiStar(t *testing.T) {
	const expect, sigmap, phi = 1.1528546895801364, 0.059995984286488495, 1.1512924985234674

	phistar := phiStar(sigmap, phi)

	if math.Abs(phistar-expect) > epsilon {
		t.Errorf("phiStar(⋯) = %v, want %v", phistar, expect)
	}
}

func TestNewRating(t *testing.T) {
	const phistar, mu, v = 1.1528546895801364, 0.0, 1.7789770897239976

	os := []Opponent{
		Opponent{1400, 30, 1},
		Opponent{1550, 100, 0},
		Opponent{1700, 300, 0}}

	scaled := scaleOpponents(mu, os)

	mup, phip := newRating(phistar, mu, v, scaled)

	if math.Abs(mup - -0.20694096667525494) > epsilon {
		t.Errorf("newRating for mup")
	}

	if math.Abs(phip-0.8721991881307343) > epsilon {
		t.Errorf("newRating for phip")
	}

}

func TestUnscale(t *testing.T) {
	const mup, phip, r1, rd1 = -0.20694096667525494, 0.8721991881307343, 1464.0506705393013, 151.51652412385727

	r, rd := Unscale(mup, phip)

	if math.Abs(r-r1) > epsilon {
		t.Errorf("unscale for r")
	}

	if math.Abs(rd-rd1) > epsilon {
		t.Errorf("unscale for rd")
	}
}

func BenchmarkRate(b *testing.B) {
	p := Player{1500, 200, 0.06}
	os := []Opponent{
		Opponent{1400, 30, 1},
		Opponent{1550, 100, 0},
		Opponent{1700, 300, 0}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Rate(os)
	}
}
