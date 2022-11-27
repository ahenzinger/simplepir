package pir

import "math"
import "strings"
import "strconv"
import "fmt"
import _ "embed"

//go:embed params.csv
var lwe_params string

type Params struct {
	N     uint64  // LWE secret dimension
	Sigma float64 // LWE error distribution stddev

	L uint64 // DB height
	M uint64 // DB width

	Logq uint64 // (logarithm of) ciphertext modulus
	P    uint64 // plaintext modulus
}

func (p *Params) Delta() uint64 {
	return (1 << p.Logq) / (p.P)
}

func (p *Params) delta() uint64 {
	return uint64(math.Ceil(float64(p.Logq) / math.Log2(float64(p.P))))
}

func (p *Params) Round(x uint64) uint64 {
	Delta := p.Delta()
	v := (x + Delta/2) / Delta
	return v % p.P
}

func (p *Params) PickParams(doublepir bool, samples ...uint64) {
	if p.N == 0 || p.Logq == 0 {
		panic("Need to specify n and q!")
	}

	num_samples := uint64(0)
	for _, ns := range samples {
		if ns > num_samples {
			num_samples = ns
		}
	}

	lines := strings.Split(lwe_params, "\n")
	for _, l := range lines[1:] {
		line := strings.Split(l, ",")
		logn, _ := strconv.ParseUint(line[0], 10, 64)
		logm, _ := strconv.ParseUint(line[1], 10, 64)
		logq, _ := strconv.ParseUint(line[2], 10, 64)

		if (p.N == uint64(1<<logn)) &&
			(num_samples <= uint64(1<<logm)) &&
			(p.Logq == uint64(logq)) {
			sigma, _ := strconv.ParseFloat(line[3], 64)
			p.Sigma = sigma

			if doublepir {
				mod, _ := strconv.ParseUint(line[6], 10, 64)
				p.P = mod
			} else {
				mod, _ := strconv.ParseUint(line[5], 10, 64)
				p.P = mod
			}

			if sigma == 0.0 || p.P == 0 {
				panic("Params invalid!")
			}

			return
		}
	}

	fmt.Printf("Searched for %d, %d-by-%d, %d,\n", p.N, p.L, p.M, p.Logq)
	panic("No suitable params known!")
}

func (p *Params) PrintParams() {
	fmt.Printf("Working with: n=%d; db size=2^%d (l=%d, m=%d); logq=%d; p=%d; sigma=%f\n",
		p.N, int(math.Log2(float64(p.L))+math.Log2(float64(p.M))), p.L, p.M, p.Logq,
		p.P, p.Sigma)
}
