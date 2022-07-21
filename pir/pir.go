package main

// Defines the interface for PIR with preprocessing schemes

type PIR interface {
	Name() string

	PickParams(N, d, n, logq uint64) Params

	GetBW(info DBinfo, p Params)

	Init(info DBinfo, p Params) State

	Setup(DB *Database, shared State, p Params) (State, Msg)
	FakeSetup(DB *Database, p Params) (State, float64) // used for benchmarking online phase

	Query(i uint64, shared State, p Params, info DBinfo) (State, Msg)

	Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg

	Recover(i uint64, batch_index uint64, offline Msg, answer Msg, client State,
		p Params, info DBinfo) uint64

	Reset(DB *Database, p Params) // reset DB to its correct state, if modified during execution
}
