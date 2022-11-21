package pir

// #cgo CFLAGS: -O3 -march=native -msse4.1 -maes -mavx2 -mavx
// #include "pir.h"
import "C"
import "fmt"
import "math/big"

type Matrix struct {
	rows uint64
	cols uint64
	data []C.Elem
}

func (m *Matrix) size() uint64 {
	return m.rows * m.cols
}

func (m *Matrix) NumRows() uint64 {
	return m.rows
}

func (m *Matrix) NumCols() uint64 {
	return m.cols
}

func (m *Matrix) AppendZeros(n uint64) {
	m.Concat(MatrixZeros(n, 1))
}

func MatrixNew(rows uint64, cols uint64) *Matrix {
	out := new(Matrix)
	out.rows = rows
	out.cols = cols
	out.data = make([]C.Elem, rows*cols)
	return out
}

func MatrixNewNoAlloc(rows uint64, cols uint64) *Matrix {
	out := new(Matrix)
	out.rows = rows
	out.cols = cols
	return out
}

func MatrixRand(rows uint64, cols uint64, logmod uint64, mod uint64) *Matrix {
	out := MatrixNew(rows, cols)
	m := big.NewInt(int64(mod))
	if mod == 0 {
		m = big.NewInt(1 << logmod)
	}
	for i := 0; i < len(out.data); i++ {
		out.data[i] = C.Elem(RandInt(m).Uint64())
	}
	return out
}

func MatrixZeros(rows uint64, cols uint64) *Matrix {
	out := MatrixNew(rows, cols)
	for i := 0; i < len(out.data); i++ {
		out.data[i] = C.Elem(0)
	}
	return out
}

func MatrixGaussian(rows, cols uint64) *Matrix {
	out := MatrixNew(rows, cols)
	for i := 0; i < len(out.data); i++ {
		out.data[i] = C.Elem(GaussSample())
	}
	return out
}

func (m *Matrix) ReduceMod(p uint64) {
	mod := C.Elem(p)
	for i := 0; i < len(m.data); i++ {
		m.data[i] = m.data[i] % mod
	}
}

func (m *Matrix) Get(i, j uint64) uint64 {
	if i >= m.rows {
		panic("Too many rows!")
	}
	if j >= m.cols {
		panic("Too many cols!")
	}
	return uint64(m.data[i*m.cols+j])
}

func (m *Matrix) Set(val, i, j uint64) {
	if i >= m.rows {
		panic("Too many rows!")
	}
	if j >= m.cols {
		panic("Too many cols!")
	}
	m.data[i*m.cols+j] = C.Elem(val)
}

func (a *Matrix) MatrixAdd(b *Matrix) {
	if (a.cols != b.cols) || (a.rows != b.rows) {
		fmt.Printf("%d-by-%d vs. %d-by-%d\n", a.rows, a.cols, b.rows, b.cols)
		panic("Dimension mismatch")
	}
	for i := uint64(0); i < a.cols*a.rows; i++ {
		a.data[i] += b.data[i]
	}
}

func (a *Matrix) Add(val uint64) {
	v := C.Elem(val)
	for i := uint64(0); i < a.cols*a.rows; i++ {
		a.data[i] += v
	}
}

func (a *Matrix) MatrixSub(b *Matrix) {
	if (a.cols != b.cols) || (a.rows != b.rows) {
		fmt.Printf("%d-by-%d vs. %d-by-%d\n", a.rows, a.cols, b.rows, b.cols)
		panic("Dimension mismatch")
	}
	for i := uint64(0); i < a.cols*a.rows; i++ {
		a.data[i] -= b.data[i]
	}
}

func (a *Matrix) Sub(val uint64) {
	v := C.Elem(val)
	for i := uint64(0); i < a.cols*a.rows; i++ {
		a.data[i] -= v
	}
}

func MatrixMul(a *Matrix, b *Matrix) *Matrix {
	if b.cols == 1 {
		return MatrixMulVec(a, b)
	}
	if a.cols != b.rows {
		fmt.Printf("%d-by-%d vs. %d-by-%d\n", a.rows, a.cols, b.rows, b.cols)
		panic("Dimension mismatch")
	}

	out := MatrixZeros(a.rows, b.cols)

	outPtr := (*C.Elem)(&out.data[0])
	aPtr := (*C.Elem)(&a.data[0])
	bPtr := (*C.Elem)(&b.data[0])
	aRows := C.size_t(a.rows)
	aCols := C.size_t(a.cols)
	bCols := C.size_t(b.cols)

	C.matMul(outPtr, aPtr, bPtr, aRows, aCols, bCols)

	return out
}

func MatrixMulTransposedPacked(a *Matrix, b *Matrix, basis, compression uint64) *Matrix {
        fmt.Printf("%d-by-%d vs. %d-by-%d\n", a.rows, a.cols, b.cols, b.rows)
        if compression != 3 && basis != 10 {
                panic("Must use hard-coded values!")
        }

        out := MatrixZeros(a.rows, b.rows)

        outPtr := (*C.Elem)(&out.data[0])
        aPtr := (*C.Elem)(&a.data[0])
        bPtr := (*C.Elem)(&b.data[0])
        aRows := C.size_t(a.rows)
	aCols := C.size_t(a.cols)
        bRows := C.size_t(b.rows)
        bCols := C.size_t(b.cols)

        C.matMulTransposedPacked(outPtr, aPtr, bPtr, aRows, aCols, bRows, bCols)

	return out
}

func MatrixMulVec(a *Matrix, b *Matrix) *Matrix {
	if (a.cols != b.rows) && (a.cols+1 != b.rows) && (a.cols+2 != b.rows) { // do not require eact match because of DB compression
		fmt.Printf("%d-by-%d vs. %d-by-%d\n", a.rows, a.cols, b.rows, b.cols)
		panic("Dimension mismatch")
	}
	if b.cols != 1 {
		panic("Second argument is not a vector")
	}

	out := MatrixNew(a.rows, 1)

	outPtr := (*C.Elem)(&out.data[0])
	aPtr := (*C.Elem)(&a.data[0])
	bPtr := (*C.Elem)(&b.data[0])
	aRows := C.size_t(a.rows)
	aCols := C.size_t(a.cols)

	C.matMulVec(outPtr, aPtr, bPtr, aRows, aCols)

	return out
}

func MatrixMulVecPacked(a *Matrix, b *Matrix, basis, compression uint64) *Matrix {
	if a.cols*compression != b.rows {
		fmt.Printf("%d-by-%d vs. %d-by-%d\n", a.rows, a.cols, b.rows, b.cols)
		panic("Dimension mismatch")
	}
	if b.cols != 1 {
		panic("Second argument is not a vector")
	}
	if compression != 3 && basis != 10 {
		panic("Must use hard-coded values!")
	}

	out := MatrixNew(a.rows+8, 1)

	outPtr := (*C.Elem)(&out.data[0])
	aPtr := (*C.Elem)(&a.data[0])
	bPtr := (*C.Elem)(&b.data[0])

	C.matMulVecPacked(outPtr, aPtr, bPtr, C.size_t(a.rows), C.size_t(a.cols))
	out.DropLastRows(8)

	return out
}

func (m *Matrix) Transpose() {
	if m.cols == 1 {
		m.cols = m.rows
		m.rows = 1
		return
	}
	if m.rows == 1 {
		m.rows = m.cols
		m.cols = 1
		return
	}

	out := MatrixNew(m.cols, m.rows)

	outPtr := (*C.Elem)(&out.data[0])
	Ptr := (*C.Elem)(&m.data[0])
	rows := C.size_t(m.rows)
	cols := C.size_t(m.cols)

	C.transpose(outPtr, Ptr, rows, cols)

	m.cols = out.cols
	m.rows = out.rows
	m.data = out.data
}

func (a *Matrix) Concat(b *Matrix) {
	if a.cols == 0 && a.rows == 0 {
		a.cols = b.cols
		a.rows = b.rows
		a.data = b.data
		return
	}

	if a.cols != b.cols {
		fmt.Printf("%d-by-%d vs. %d-by-%d\n", a.rows, a.cols, b.rows, b.cols)
		panic("Dimension mismatch")
	}

	a.rows += b.rows
	a.data = append(a.data, b.data...)
}

// Represent each element in the database with 'delta' elements in Z_'mod'.
// Then, map the database elements from [0, mod] to [-mod/2, mod/2].
func (m *Matrix) Expand(mod uint64, delta uint64) {
	n := MatrixNew(m.rows*delta, m.cols)
	modulus := C.Elem(mod)

	for i := uint64(0); i < m.rows; i++ {
		for j := uint64(0); j < m.cols; j++ {
			val := m.data[i*m.cols+j]
			for f := uint64(0); f < delta; f++ {
				new_val := val % modulus
				n.data[(i*delta+f)*m.cols+j] = new_val - modulus/2
				val /= modulus
			}
		}
	}

	m.cols = n.cols
	m.rows = n.rows
	m.data = n.data
}

func (m *Matrix) TransposeAndExpandAndConcatColsAndSquish(mod, delta, concat, basis, d uint64) {
        if m.rows % concat != 0 {
                panic("Bad input!")
        }

        n := MatrixZeros(m.cols*delta*concat, (m.rows/concat+d-1)/d)

        for j := uint64(0); j < m.rows; j++ {
                for i := uint64(0); i < m.cols; i++ {
                        val := uint64(m.data[i+j*m.cols])
                        for f := uint64(0); f < delta; f++ {
                                new_val := val % mod
                                r := (i*delta+f) + m.cols*delta*(j % concat)
                                c := j / concat
                                n.data[r*n.cols+c/d] += C.Elem(new_val << (basis * (c%d)))
                                val /= mod
                        }
                }
        }

        m.cols = n.cols
        m.rows = n.rows
        m.data = n.data
}

// Computes the inverse operations of Expand(.)
func (m *Matrix) Contract(mod uint64, delta uint64) {
	n := MatrixZeros(m.rows/delta, m.cols)

	for i := uint64(0); i < n.rows; i++ {
		for j := uint64(0); j < n.cols; j++ {
			var vals []uint64
			for f := uint64(0); f < delta; f++ {
				new_val := uint64(m.data[(i*delta+f)*m.cols+j])
				vals = append(vals, (new_val+mod/2)%mod)
			}
			n.data[i*m.cols+j] += C.Elem(Reconstruct_from_base_p(mod, vals))
		}
	}

	m.cols = n.cols
	m.rows = n.rows
	m.data = n.data
}

// Squishes the matrix by representing each group of 'delta' consecutive value
// as a single database element, where each value uses 'basis' bits.
func (m *Matrix) Squish(basis, delta uint64) {
	n := MatrixZeros(m.rows, (m.cols+delta-1)/delta)

	for i := uint64(0); i < n.rows; i++ {
		for j := uint64(0); j < n.cols; j++ {
			for k := uint64(0); k < delta; k++ {
				if delta*j+k < m.cols {
					val := m.Get(i, delta*j+k)
					n.data[i*n.cols+j] += C.Elem(val << (k * basis))
				}
			}
		}
	}

	m.cols = n.cols
	m.rows = n.rows
	m.data = n.data
}

// Computes the inverse operation of Squish(.)
func (m *Matrix) Unsquish(basis, delta, cols uint64) {
	n := MatrixZeros(m.rows, cols)
	mask := uint64((1 << basis) - 1)

	for i := uint64(0); i < m.rows; i++ {
		for j := uint64(0); j < m.cols; j++ {
			for k := uint64(0); k < delta; k++ {
				if j*delta+k < cols {
					n.data[i*n.cols+j*delta+k] = C.Elem(((m.Get(i, j)) >> (k * basis)) & mask)
				}
			}
		}
	}

	m.cols = n.cols
	m.rows = n.rows
	m.data = n.data
}

func (m *Matrix) Round(p Params) {
	for i := uint64(0); i < m.rows*m.cols; i++ {
		m.data[i] = C.Elem(p.Round(uint64(m.data[i])))
	}
}

func (m *Matrix) DropLastRows(n uint64) {
	m.rows -= n
	m.data = m.data[:(m.rows * m.cols)]
}

func (m *Matrix) Column(i uint64) *Matrix {
	if m.cols == 1 {
		return m
	}

	col := MatrixNew(m.rows, 1)
	for j := uint64(0); j < m.rows; j++ {
		col.data[j] = m.data[j*m.cols+i]
	}
	return col
}

func (m *Matrix) Rows(offset, num_rows uint64) *Matrix {
	if (offset == 0) && (num_rows == m.rows) {
		return m
	}

	if offset > m.rows {
		panic("Asking for bad offset!")
	}

	if offset+num_rows <= m.rows {
		m2 := MatrixNewNoAlloc(num_rows, m.cols)
		m2.data = m.data[(offset * m.cols) : (offset+num_rows)*m.cols]
		return m2
	}

	m2 := MatrixNewNoAlloc(m.rows-offset, m.cols)
	m2.data = m.data[(offset * m.cols) : (m.rows)*m.cols]

	return m2
}

func (m *Matrix) RowsDeepCopy(offset, num_rows uint64) *Matrix {
	if offset+num_rows > m.rows {
		panic("Requesting too many rows")
	}

	if offset+num_rows <= m.rows {
		m2 := MatrixNew(num_rows, m.cols)
		copy(m2.data, m.data[(offset*m.cols):((offset+num_rows)*m.cols)])
		return m2
	}

	m2 := MatrixNew(m.rows-offset, m.cols)
	copy(m2.data, m.data[(offset*m.cols):(m.rows)*m.cols])
	return m2
}

func (m *Matrix) ConcatCols(n uint64) {
	if n == 1 {
		return
	}

	fmt.Printf("Running concat cols on matrix of dims %d-by-%d with n=%d\n",
		m.rows, m.cols, n)

	if m.cols%n != 0 {
		panic("n does not divide num cols")
	}

	m2 := MatrixNew(m.rows*n, m.cols/n)
	for i := uint64(0); i < m.rows; i++ {
		for j := uint64(0); j < m.cols; j++ {
			col := j / n
			row := i + m.rows*(j%n)
			m2.data[row*m2.cols+col] = m.data[i*m.cols+j]
		}
	}

	m.cols = m2.cols
	m.rows = m2.rows
	m.data = m2.data
}

func (m *Matrix) Dim() {
	fmt.Printf("Dims: %d-by-%d\n", m.rows, m.cols)
}

func (m *Matrix) Print() {
	fmt.Printf("%d-by-%d matrix:\n", m.rows, m.cols)
	for i := uint64(0); i < m.rows; i++ {
		for j := uint64(0); j < m.cols; j++ {
			fmt.Printf("%d ", m.data[i*m.cols+j])
		}
		fmt.Printf("\n")
	}
}

func (m *Matrix) PrintStart() {
        fmt.Printf("%d-by-%d matrix:\n", m.rows, m.cols)
        for i := uint64(0); i < 2; i++ {
                for j := uint64(0); j < 2; j++ {
                        fmt.Printf("%d ", m.data[i*m.cols+j])
                }
                fmt.Printf("\n")
        }
}
