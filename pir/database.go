package main

import "math"
import "fmt"

type DBinfo struct {
  N          uint64   // number of DB rows.
  row_length uint64   // number of bits per DB row/entry.

  packing    uint64   // number of entries per DB elem, if DB elem size > entry size.
  ne         uint64   // number of DB elems per entry, if entry size > DB elem size.

  x          uint64   // tunable param that governs communication,
                      // must be in range [1, ne] and must be a divisor of ne;
		      // represents the number of times the scheme is repeated.
  p          uint64   // LWE plaintext modulus, i.e. DB elem size.
  logq       uint64   // LWE modulus.

  basis      uint64   // for DB compression
  squishing  uint64
  cols       uint64
}

type Database struct {
  info       DBinfo
  data       *Matrix
}


func (DB *Database) Squish() {
  fmt.Printf("Original DB dims: ")
  DB.data.Dim()

  DB.info.basis = 10//uint64(math.Ceil(math.Log2(float64(DB.info.p))))
  DB.info.squishing = 3//DB.info.logq / DB.info.basis
  DB.info.cols = DB.data.cols
  DB.data.Squish(DB.info.basis, DB.info.squishing)

  fmt.Printf("After squishing, with compression factor %d: ", DB.info.squishing)
  DB.data.Dim()

  // Check that params allow for this compression
  if uint64(math.Ceil(math.Log2(float64(DB.info.p))))*DB.info.squishing > DB.info.logq {
    panic("Bad params")
  }
}

func (DB *Database) Unsquish() {
  DB.data.Unsquish(DB.info.basis, DB.info.squishing, DB.info.cols)
}

// For a structured DB: Stack items in same row below each other,
// so can retrieve them in 1 round of PIR.

func ReconstructElem(vals []uint64, index uint64, info DBinfo) uint64 {
  q := uint64(1 << info.logq)

  for i, _ := range vals {
    vals[i] = (vals[i] + info.p/2) % q
    vals[i] = vals[i] % info.p
  }

  val := Reconstruct_from_base_p(info.p, vals)

  if info.packing > 0 {
    val = Base_p((1 << info.row_length), val, index % info.packing)
  }

  return val
}

func (DB *Database) GetElem(i uint64) uint64 {
  if i >= DB.info.N {
    panic("Index out of range")
  }

  col := i % DB.data.cols
  row := i / DB.data.cols

  if DB.info.packing > 0 {
    new_i := i / DB.info.packing
    col = new_i % DB.data.cols
    row = new_i / DB.data.cols
  }

  var vals []uint64
  for j := row * DB.info.ne; j < (row+1) * DB.info.ne; j++ {
    vals = append(vals, DB.data.Get(j, col))
  }

  return ReconstructElem(vals, i, DB.info)
}

// Find smallest l, m such that l*m >= N*ne and ne divides l, where ne is
// the number of entries per row determined by row_length and p.
func ApproxSquareDatabaseDims(N, row_length, p uint64) (uint64, uint64) {
  db_elems, elems_per_entry, _ := Num_DB_entries(N, row_length, p)
  l := uint64(math.Floor(math.Sqrt(float64(db_elems))))

  rem := l % elems_per_entry
  if rem != 0 {
    l += elems_per_entry - rem
  }

  m := uint64(math.Ceil(float64(db_elems)/float64(l)))

  return l, m
}

// Find smallest l, m such that l*m >= N*ne and ne divides l, where ne is
// the number of entries per row determined by row_length and p, and m >=
// lower_bound_m.
func ApproxDatabaseDims(N, row_length, p, lower_bound_m uint64) (uint64, uint64) {
  l, m := ApproxSquareDatabaseDims(N, row_length, p)
  if m >= lower_bound_m {
    return l, m
  }

  m = lower_bound_m
  db_elems, elems_per_entry, _ := Num_DB_entries(N, row_length, p)
  l = uint64(math.Ceil(float64(db_elems)/float64(m)))

  rem := l % elems_per_entry
  if rem != 0 {
    l += elems_per_entry - rem
  }

  return l, m
}

func SetupDB(N, row_length uint64, p *Params) (*Database) {
  if (N == 0) || (row_length == 0) {
    panic("Empty database!")
  }

  D := new(Database)

  D.info.N = N
  D.info.row_length = row_length
  D.info.p = p.p
  D.info.logq = p.logq

  db_elems, elems_per_entry, entries_per_elem := Num_DB_entries(N, row_length, p.p)
  D.info.ne = elems_per_entry
  D.info.x = D.info.ne//uint64(math.Sqrt(float64(D.info.ne)))
  D.info.packing = entries_per_elem

  for ; D.info.ne % D.info.x != 0; {
    D.info.x += 1
  }

  D.info.basis = 0
  D.info.squishing = 0

  //fmt.Printf("x is %d; ne/x is %d; ne is %d\n", D.info.x, D.info.ne/D.info.x, D.info.ne)
  fmt.Printf("Total packed DB size is ~%f MB\n",
             float64(p.l * p.m) * math.Log2(float64(p.p)) / (1024.0 * 1024.0 * 8.0))

  if db_elems > p.l * p.m {
    panic("Params and database size don't match")
  }

  if p.l % D.info.ne != 0 {
    panic("Number of DB elems per entry must divide DB height")
  }

  return D
}

func MakeRandomDB(N, row_length uint64, p *Params ) (*Database) {
  D := SetupDB(N, row_length, p)

  D.data = MatrixRand(p.l, p.m, 0, p.p)

  // Map DB elems to [-p/2; p/2]
  D.data.Sub(p.p/2)

  return D
}

func MakeDB(N, row_length uint64, p *Params, vals []uint64) (*Database) {
  D := SetupDB(N, row_length, p)
  D.data = MatrixZeros(p.l, p.m)

  if uint64(len(vals)) != N {
    panic("Bad input DB")
  }

  if D.info.packing > 0 {
    // Pack multiple DB elems into each Z_p elem
    at := uint64(0)
    cur := uint64(0)
    coeff := uint64(1)
    for i, elem := range vals {
      cur += (elem * coeff)
      coeff *= (1 << row_length)
      if ((i+1) % int(D.info.packing) == 0) || (i == len(vals)-1) {
        D.data.Set(cur, at/p.m, at % p.m)
	at += 1
	cur = 0
	coeff = 1
      }
    }
  } else {
    // Use multiple Z_p elems to represent each DB elem
    for i, elem := range vals {
      for j := uint64(0); j < D.info.ne; j++ {
        D.data.Set(Base_p(D.info.p, elem, j), (uint64(i)/p.m)*D.info.ne + j, uint64(i) % p.m)
      }
    }
  }

  // Map DB elems to [-p/2; p/2]
  D.data.Sub(p.p/2)

  return D
}
