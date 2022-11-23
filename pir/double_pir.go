package pir

// #cgo CFLAGS: -O3 -march=native -msse4.1 -maes -mavx2 -mavx
// #include "pir.h"
import "C"
import "fmt"

type DoublePIR struct{}

// Offline download: matrix H2
// Online query: matrices q1, q2
// Online download: matrices h1, a2, h2

// Server state: matrix H1
// Client state: matrices secret1, secret2
// Shared state: matrices A1, A2

// Ratio between first-level DB and second-level DB
const COMP_RATIO = uint64(64)

func (pi *DoublePIR) Name() string {
	return "DoublePIR"
}

func (pi *DoublePIR) PickParams(N, d, n, logq uint64) Params {
	good_p := Params{}
	found := false

	// Iteratively refine p and DB dims, until find tight values
	for mod_p := uint64(2); ; mod_p += 1 {
		l, m := ApproxDatabaseDims(N, d, mod_p, COMP_RATIO*n)

		p := Params{
			n:    n,
			logq: logq,
			l:    l,
			m:    m,
		}
		p.PickParams(true, l, m)

		if p.p < mod_p {
			if !found {
				panic("Error; should not happen")
			}
			good_p.PrintParams()
			return good_p
		}

		good_p = p
		found = true
	}

	panic("Cannot be reached")
	return Params{}
}

func (pi *DoublePIR) PickParamsGivenDimensions(l, m, n, logq uint64) Params {
	p := Params{
		n:    n,
		logq: logq,
		l:    l,
                m:    m,
	}
        p.PickParams(true, l, m)
        return p
}

func (pi *DoublePIR) GetBW(info DBinfo, p Params) {
	offline_download := float64(p.delta()*info.x*p.n*p.n*p.logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	online_upload := float64(p.m*p.logq+info.ne/info.x*p.l/info.x*p.logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline upload: %d KB\n", uint64(online_upload))

	online_download := float64(p.delta()*info.x*p.n*p.logq+p.delta()*p.n*info.ne*p.logq+p.delta()*info.ne*p.logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline download: %d KB\n", uint64(online_download))
}

func (pi *DoublePIR) Init(info DBinfo, p Params) State {
	A1 := MatrixRand(p.m, p.n, p.logq, 0)
	A2 := MatrixRand(p.l/info.x, p.n, p.logq, 0)

	return MakeState(A1, A2)
}

func (pi *DoublePIR) InitCompressed(info DBinfo, p Params) (State, CompressedState) {
        seed := RandomPRGKey()
        bufPrgReader = NewBufPRG(NewPRG(seed))
        return pi.Init(info, p), MakeCompressedState(seed)
}

func (pi *DoublePIR) DecompressState(info DBinfo, p Params, comp CompressedState) State {
        bufPrgReader = NewBufPRG(NewPRG(comp.seed))
        return pi.Init(info, p)
}

func (pi *DoublePIR) Setup(DB *Database, shared State, p Params) (State, Msg) {
	A1 := shared.data[0]
	A2 := shared.data[1]

	H1 := MatrixMul(DB.data, A1)
	H1.Transpose()
	H1.Expand(p.p, p.delta())
	H1.ConcatCols(DB.info.x)

	H2 := MatrixMul(H1, A2)

	// pack the database more tightly, because the online computation is memory-bound
	DB.data.Add(p.p / 2)
	DB.Squish()

	H1.Add(p.p / 2)
	H1.Squish(10, 3)

	A2_copy := A2.RowsDeepCopy(0, A2.rows) // deep copy whole matrix
	if A2_copy.rows % 3 != 0 {
                A2_copy.Concat(MatrixZeros(3-(A2_copy.rows%3), A2_copy.cols))
        }
	A2_copy.Transpose()

	return MakeState(H1, A2_copy), MakeMsg(H2)
}

func (pi *DoublePIR) FakeSetup(DB *Database, p Params) (State, float64) {
	info := DB.info
	H1 := MatrixRand(p.n*p.delta()*info.x, p.l/info.x, 0, p.p)
	offline_download := float64(p.n*p.delta()*info.x*p.n*uint64(p.logq)) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	// pack the database more tightly, because the online computation is memory-bound
	DB.data.Add(p.p / 2)
	DB.Squish()

	H1.Add(p.p / 2)
	H1.Squish(10, 3)

	A2_rows := p.l/info.x
	if A2_rows % 3 != 0 {
		A2_rows += (3-(A2_rows % 3))
	}
	A2_copy := MatrixRand(p.n, A2_rows, p.logq, 0)

	return MakeState(H1, A2_copy), offline_download
}

func (pi *DoublePIR) Query(i uint64, shared State, p Params, info DBinfo) (State, Msg) {
	i1 := (i / p.m) * (info.ne / info.x)
	i2 := i % p.m

	A1 := shared.data[0]
	A2 := shared.data[1]

	secret1 := MatrixRand(p.n, 1, p.logq, 0)
	err1 := MatrixGaussian(p.m, 1)
	query1 := MatrixMul(A1, secret1)
	query1.MatrixAdd(err1)
	query1.data[i2] += C.Elem(p.Delta())

	if p.m%info.squishing != 0 {
		query1.AppendZeros(info.squishing - (p.m % info.squishing))
	}

	state := MakeState(secret1)
	msg := MakeMsg(query1)

	for j := uint64(0); j < info.ne/info.x; j++ {
		secret2 := MatrixRand(p.n, 1, p.logq, 0)
		err2 := MatrixGaussian(p.l/info.x, 1)
		query2 := MatrixMul(A2, secret2)
		query2.MatrixAdd(err2)
		query2.data[i1+j] += C.Elem(p.Delta())

		if (p.l/info.x)%info.squishing != 0 {
			query2.AppendZeros(info.squishing - ((p.l / info.x) % info.squishing))
		}

		state.data = append(state.data, secret2)
		msg.data = append(msg.data, query2)
	}

	return state, msg
}

func (pi *DoublePIR) Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg {
	H1 := server.data[0]
	A2_transpose := server.data[1]

	a1 := new(Matrix)
	num_queries := uint64(len(query.data))
	batch_sz := DB.data.rows / num_queries

	last := uint64(0)
	for batch, q := range query.data {
		q1 := q.data[0]
		if batch == int(num_queries-1) {
			batch_sz = DB.data.rows - last
		}
		a := MatrixMulVecPacked(DB.data.Rows(last, batch_sz),
			                q1, DB.info.basis, DB.info.squishing)
		a1.Concat(a)
		last += batch_sz
	}

	a1.TransposeAndExpandAndConcatColsAndSquish(p.p, p.delta(), DB.info.x, 10, 3)
        h1 := MatrixMulTransposedPacked(a1, A2_transpose, 10, 3)
	msg := MakeMsg(h1)

	for _, q := range query.data {
		for j := uint64(0); j < DB.info.ne/DB.info.x; j++ {
			q2 := q.data[1+j]
			a2 := MatrixMulVecPacked(H1, q2, 10, 3)
			h2 := MatrixMulVecPacked(a1, q2, 10, 3)

			msg.data = append(msg.data, a2)
			msg.data = append(msg.data, h2)
		}
	}

	return msg
}

func (pi *DoublePIR) Recover(i uint64, batch_index uint64, offline Msg, query Msg,
	answer Msg, shared State, client State, p Params, info DBinfo) uint64 {
	H2 := offline.data[0]
	h1 := answer.data[0].RowsDeepCopy(0, answer.data[0].rows) // deep copy whole matrix 
	secret1 := client.data[0]

	ratio := p.p/2
	val1 := uint64(0)
	for j := uint64(0); j<p.m; j++ {
		val1 += ratio*query.data[0].Get(j,0)
	}
	val1 %= (1<<p.logq)
	val1 = (1<<p.logq)-val1

	val2 := uint64(0)
	for j := uint64(0); j<p.l/info.x; j++ {
		val2 += ratio*query.data[1].Get(j,0)
	}
	val2 %= (1<<p.logq)
	val2 = (1<<p.logq)-val2

	A2 := shared.data[1]
	if (A2.cols != p.n) || (h1.cols != p.n) {
		panic("Should not happen!")
	}
	for j1 := uint64(0); j1<p.n; j1++ {
		val3 := uint64(0)
	        for j2 := uint64(0); j2<A2.rows; j2++ {
			val3 += ratio*A2.Get(j2,j1)
		}
		val3 %= (1<<p.logq)
		val3 = (1<<p.logq)-val3
		v := C.Elem(val3)
		for k := uint64(0); k<h1.rows; k++ {
                	h1.data[k*h1.cols+j1] += v
		}
	}

	offset := (info.ne / info.x * 2) * batch_index // for batching
	var vals []uint64
	for i := uint64(0); i < info.ne/info.x; i++ {
		a2 := answer.data[1+2*i+offset]
		h2 := answer.data[2+2*i+offset]
		secret2 := client.data[1+i]
		h2.Add(val2)

		for j := uint64(0); j < info.x; j++ {
			state := a2.RowsDeepCopy(j*p.n*p.delta(), p.n*p.delta())
			state.Add(val2)
			state.Concat(h2.Rows(j*p.delta(), p.delta()))

			hint := H2.RowsDeepCopy(j*p.n*p.delta(), p.n*p.delta())
			hint.Concat(h1.Rows(j*p.delta(), p.delta()))

			interm := MatrixMul(hint, secret2)
			state.MatrixSub(interm)
			state.Round(p)
			state.Contract(p.p, p.delta())

			noised := uint64(state.data[p.n]) + val1
			for l := uint64(0); l < p.n; l++ {
				noised -= uint64(secret1.data[l] * state.data[l])
				noised = noised % (1 << p.logq)
			}
			vals = append(vals, p.Round(noised))
			//fmt.Printf("Reconstructing row %d: %d\n", j+info.x*i, denoised)
		}
	}

	return ReconstructElem(vals, i, info)
}

func (pi *DoublePIR) Reset(DB *Database, p Params) {
	DB.Unsquish()
	DB.data.Sub(p.p / 2)
}
