package main

import "math"
import "encoding/csv"
import "os"
import "strconv"
import "fmt"

const filename = "params.csv"

type Params struct {
  n     uint64    // LWE secret dimension
  sigma float64   // LWE error dist stddev

  l     uint64    // DB height
  m     uint64    // DB width

  logq  uint64    // ciphertext modulus
  p  uint64       // plaintext modulus
}

func (p *Params) Delta() uint64 {
  return (1 << p.logq) / (p.p)
}

func (p *Params) delta() uint64 {
  return uint64(math.Ceil(float64(p.logq) / math.Log2(float64(p.p))))
}

func (p *Params) Round(x uint64) uint64 {
  Delta := p.Delta()
  v := (x + Delta/2) / Delta
  return v % p.p
}

func (p *Params) PickParams(doublepir bool, samples ...uint64) {
  if p.n == 0 || p.logq == 0 {
    panic("Need to specify n and q!")
  }

  num_samples := uint64(0)
  for _, ns := range samples {
    if ns > num_samples {
      num_samples = ns
    }
  }

  f, err := os.Open(filename)
  if err != nil {
    panic("Error opening csv file")
  }
  defer f.Close()

  records, err := csv.NewReader(f).ReadAll()
  if err != nil {
    fmt.Printf("%s\n", err)
    panic("Error reading csv file")
  }

  for _, line := range records[1:] {
    logn, _ := strconv.ParseUint(line[0], 10, 64)
    logm, _ := strconv.ParseUint(line[1], 10, 64)
    logq, _ := strconv.ParseUint(line[2], 10, 64)

    if (p.n == uint64(1 << logn)) &&
       (num_samples <= uint64(1 << logm)) &&
       (p.logq == uint64(logq)) {
      sigma, _ := strconv.ParseFloat(line[3], 64)
      p.sigma = sigma

      if doublepir {
	mod, _ := strconv.ParseUint(line[6], 10, 64)
        p.p = mod
      } else {
	mod, _ := strconv.ParseUint(line[5], 10, 64)
        p.p = mod
      }

      if sigma == 0.0 || p.p == 0 {
        panic("Params invalid!")
      }

      return
    }
  }

  fmt.Printf("Searched for %d, %d-by-%d, %d,\n", p.n, p.l, p.m, p.logq)
  panic("No suitable params known!")
}

func (p *Params) PrintParams() {
  fmt.Printf("Working with: n=%d; db size=2^%d (l=%d, m=%d); logq=%d; p=%d; sigma=%f\n",
             p.n, int(math.Log2(float64(p.l))+math.Log2(float64(p.m))), p.l, p.m, p.logq,
             p.p, p.sigma)
}
