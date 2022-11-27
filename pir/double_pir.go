package pir

// #cgo CFLAGS: -O3 -march=native
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
			N:    n,
			Logq: logq,
			L:    l,
			M:    m,
		}
		p.PickParams(true, l, m)

		if p.P < mod_p {
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
		N:    n,
		Logq: logq,
		L:    l,
                M:    m,
	}
        p.PickParams(true, l, m)
        return p
}

func (pi *DoublePIR) GetBW(info DBinfo, p Params) {
	offline_download := float64(p.delta()*info.X*p.N*p.N*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	online_upload := float64(p.M*p.Logq+info.Ne/info.X*p.L/info.X*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline upload: %d KB\n", uint64(online_upload))

	online_download := float64(p.delta()*info.X*p.N*p.Logq+p.delta()*p.N*info.Ne*p.Logq+p.delta()*info.Ne*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline download: %d KB\n", uint64(online_download))
}

func (pi *DoublePIR) Init(info DBinfo, p Params) State {
	A1 := MatrixRand(p.M, p.N, p.Logq, 0)
	A2 := MatrixRand(p.L/info.X, p.N, p.Logq, 0)

	return MakeState(A1, A2)
}

func (pi *DoublePIR) InitCompressed(info DBinfo, p Params) (State, CompressedState) {
        seed := RandomPRGKey()
        bufPrgReader = NewBufPRG(NewPRG(seed))
        return pi.Init(info, p), MakeCompressedState(seed)
}

func (pi *DoublePIR) DecompressState(info DBinfo, p Params, comp CompressedState) State {
        bufPrgReader = NewBufPRG(NewPRG(comp.Seed))
        return pi.Init(info, p)
}

func (pi *DoublePIR) Setup(DB *Database, shared State, p Params) (State, Msg) {
	A1 := shared.Data[0]
	A2 := shared.Data[1]

	H1 := MatrixMul(DB.Data, A1)
	H1.Transpose()
	H1.Expand(p.P, p.delta())
	H1.ConcatCols(DB.Info.X)

	H2 := MatrixMul(H1, A2)

	// pack the database more tightly, because the online computation is memory-bound
	DB.Data.Add(p.P / 2)
	DB.Squish()

	H1.Add(p.P / 2)
	H1.Squish(10, 3)

	A2_copy := A2.RowsDeepCopy(0, A2.Rows) // deep copy whole matrix
	if A2_copy.Rows % 3 != 0 {
                A2_copy.Concat(MatrixZeros(3-(A2_copy.Rows%3), A2_copy.Cols))
        }
	A2_copy.Transpose()

	return MakeState(H1, A2_copy), MakeMsg(H2)
}

func (pi *DoublePIR) FakeSetup(DB *Database, p Params) (State, float64) {
	info := DB.Info
	H1 := MatrixRand(p.N*p.delta()*info.X, p.L/info.X, 0, p.P)
	offline_download := float64(p.N*p.delta()*info.X*p.N*uint64(p.Logq)) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	// pack the database more tightly, because the online computation is memory-bound
	DB.Data.Add(p.P / 2)
	DB.Squish()

	H1.Add(p.P / 2)
	H1.Squish(10, 3)

	A2_rows := p.L/info.X
	if A2_rows % 3 != 0 {
		A2_rows += (3-(A2_rows % 3))
	}
	A2_copy := MatrixRand(p.N, A2_rows, p.Logq, 0)

	return MakeState(H1, A2_copy), offline_download
}

func (pi *DoublePIR) Query(i uint64, shared State, p Params, info DBinfo) (State, Msg) {
	i1 := (i / p.M) * (info.Ne / info.X)
	i2 := i % p.M

	A1 := shared.Data[0]
	A2 := shared.Data[1]

	secret1 := MatrixRand(p.N, 1, p.Logq, 0)
	err1 := MatrixGaussian(p.M, 1)
	query1 := MatrixMul(A1, secret1)
	query1.MatrixAdd(err1)
	query1.Data[i2] += C.Elem(p.Delta())

	if p.M%info.Squishing != 0 {
		query1.AppendZeros(info.Squishing - (p.M % info.Squishing))
	}

	state := MakeState(secret1)
	msg := MakeMsg(query1)

	for j := uint64(0); j < info.Ne/info.X; j++ {
		secret2 := MatrixRand(p.N, 1, p.Logq, 0)
		err2 := MatrixGaussian(p.L/info.X, 1)
		query2 := MatrixMul(A2, secret2)
		query2.MatrixAdd(err2)
		query2.Data[i1+j] += C.Elem(p.Delta())

		if (p.L/info.X)%info.Squishing != 0 {
			query2.AppendZeros(info.Squishing - ((p.L / info.X) % info.Squishing))
		}

		state.Data = append(state.Data, secret2)
		msg.Data = append(msg.Data, query2)
	}

	return state, msg
}

func (pi *DoublePIR) Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg {
	H1 := server.Data[0]
	A2_transpose := server.Data[1]

	a1 := new(Matrix)
	num_queries := uint64(len(query.Data))
	batch_sz := DB.Data.Rows / num_queries

	last := uint64(0)
	for batch, q := range query.Data {
		q1 := q.Data[0]
		if batch == int(num_queries-1) {
			batch_sz = DB.Data.Rows - last
		}
		a := MatrixMulVecPacked(DB.Data.SelectRows(last, batch_sz),
			                q1, DB.Info.Basis, DB.Info.Squishing)
		a1.Concat(a)
		last += batch_sz
	}

	a1.TransposeAndExpandAndConcatColsAndSquish(p.P, p.delta(), DB.Info.X, 10, 3)
        h1 := MatrixMulTransposedPacked(a1, A2_transpose, 10, 3)
	msg := MakeMsg(h1)

	for _, q := range query.Data {
		for j := uint64(0); j < DB.Info.Ne/DB.Info.X; j++ {
			q2 := q.Data[1+j]
			a2 := MatrixMulVecPacked(H1, q2, 10, 3)
			h2 := MatrixMulVecPacked(a1, q2, 10, 3)

			msg.Data = append(msg.Data, a2)
			msg.Data = append(msg.Data, h2)
		}
	}

	return msg
}

func (pi *DoublePIR) Recover(i uint64, batch_index uint64, offline Msg, query Msg,
	answer Msg, shared State, client State, p Params, info DBinfo) uint64 {
	H2 := offline.Data[0]
	h1 := answer.Data[0].RowsDeepCopy(0, answer.Data[0].Rows) // deep copy whole matrix 
	secret1 := client.Data[0]

	ratio := p.P/2
	val1 := uint64(0)
	for j := uint64(0); j<p.M; j++ {
		val1 += ratio*query.Data[0].Get(j,0)
	}
	val1 %= (1<<p.Logq)
	val1 = (1<<p.Logq)-val1

	val2 := uint64(0)
	for j := uint64(0); j<p.L/info.X; j++ {
		val2 += ratio*query.Data[1].Get(j,0)
	}
	val2 %= (1<<p.Logq)
	val2 = (1<<p.Logq)-val2

	A2 := shared.Data[1]
	if (A2.Cols != p.N) || (h1.Cols != p.N) {
		panic("Should not happen!")
	}
	for j1 := uint64(0); j1<p.N; j1++ {
		val3 := uint64(0)
	        for j2 := uint64(0); j2<A2.Rows; j2++ {
			val3 += ratio*A2.Get(j2,j1)
		}
		val3 %= (1<<p.Logq)
		val3 = (1<<p.Logq)-val3
		v := C.Elem(val3)
		for k := uint64(0); k<h1.Rows; k++ {
                	h1.Data[k*h1.Cols+j1] += v
		}
	}

	offset := (info.Ne / info.X * 2) * batch_index // for batching
	var vals []uint64
	for i := uint64(0); i < info.Ne/info.X; i++ {
		a2 := answer.Data[1+2*i+offset]
		h2 := answer.Data[2+2*i+offset]
		secret2 := client.Data[1+i]
		h2.Add(val2)

		for j := uint64(0); j < info.X; j++ {
			state := a2.RowsDeepCopy(j*p.N*p.delta(), p.N*p.delta())
			state.Add(val2)
			state.Concat(h2.SelectRows(j*p.delta(), p.delta()))

			hint := H2.RowsDeepCopy(j*p.N*p.delta(), p.N*p.delta())
			hint.Concat(h1.SelectRows(j*p.delta(), p.delta()))

			interm := MatrixMul(hint, secret2)
			state.MatrixSub(interm)
			state.Round(p)
			state.Contract(p.P, p.delta())

			noised := uint64(state.Data[p.N]) + val1
			for l := uint64(0); l < p.N; l++ {
				noised -= uint64(secret1.Data[l] * state.Data[l])
				noised = noised % (1 << p.Logq)
			}
			vals = append(vals, p.Round(noised))
			//fmt.Printf("Reconstructing row %d: %d\n", j+info.X*i, denoised)
		}
	}

	return ReconstructElem(vals, i, info)
}

func (pi *DoublePIR) Reset(DB *Database, p Params) {
	DB.Unsquish()
	DB.Data.Sub(p.P / 2)
}
