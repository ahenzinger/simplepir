package main

import (
  "fmt"
  "testing"
  "time"
  "os"
  "runtime/pprof"
  "encoding/csv"
  "math"
  "strconv"
  "runtime"
  "runtime/debug"
)

const LOGQ = uint64(32)

const SEC_PARAM = uint64(1 << 10)


// Run PIR online phase, with a random setup (to skip the offline phase).
// Gives accurate bandwidth and online time measurements.
func RunFakePIR(pi PIR, DB *Database, p Params, i []uint64, f *os.File, profile bool) (float64, float64, float64, float64) {
  fmt.Printf("Executing %s\n", pi.Name())
  debug.SetGCPercent(-1)

  shared_state := pi.Init(DB.info, p)

  fmt.Println("Setup...")
  server_state, bw := pi.FakeSetup(DB, p)
  offline_comm := bw
  runtime.GC()

  fmt.Println("Building query...")
  start := time.Now()
  var query MsgSlice
  for index, _ := range i {
    _, q := pi.Query(i[index], shared_state, p, DB.info)
    query.data = append(query.data, q)
  }
  printTime(start)
  fmt.Printf("\t\tOnline upload: %f KB\n",
             float64(query.size() * uint64(p.logq)) / float64(8 * 1024))
  bw += float64(query.size() * uint64(p.logq) / (8.0 * 1024.0))
  online_comm := float64(query.size() * uint64(p.logq) / (8.0 * 1024.0))
  runtime.GC()

  fmt.Println("Answering query...")
  if(profile) {
    pprof.StartCPUProfile(f)
  }
  start = time.Now()
  answer := pi.Answer(DB, query, server_state, shared_state, p)
  elapsed := printTime(start)
  if (profile) {
    pprof.StopCPUProfile()
  }
  rate := printRate(p, elapsed, len(i))
  fmt.Printf("\t\tOnline download: %f KB\n",
             float64(answer.size() * uint64(p.logq)) / float64(8 * 1024))
  bw += float64(answer.size() * uint64(p.logq) / (8.0 * 1024.0))
  online_comm += float64(answer.size() * uint64(p.logq) / (8.0 * 1024.0))

  runtime.GC()
  debug.SetGCPercent(100)
  pi.Reset(DB, p);

  if offline_comm + online_comm != bw {
    panic("Should not happen!")
  }
  return rate, bw, offline_comm, online_comm
}

// Run full PIR scheme (offline + online phases).
func RunPIR(pi PIR, DB *Database, p Params, i []uint64) (float64, float64) {
  fmt.Printf("Executing %s\n", pi.Name())
  debug.SetGCPercent(-1)

  num_queries := uint64(len(i))
  if DB.data.rows/num_queries < DB.info.ne {
    panic("Too many queries to handle!")
  }
  batch_sz := DB.data.rows / (DB.info.ne * num_queries) * DB.data.cols
  bw := float64(0)

  shared_state := pi.Init(DB.info, p)

  fmt.Println("Setup...")
  start := time.Now()
  server_state, offline_download := pi.Setup(DB, shared_state, p)
  printTime(start)
  fmt.Printf("\t\tOffline download: %f KB\n",
             float64(offline_download.size() * uint64(p.logq)) / float64(8 * 1024))
  bw += float64(offline_download.size() * uint64(p.logq) / (8.0 * 1024.0))
  runtime.GC()

  fmt.Println("Building query...")
  start = time.Now()
  var client_state []State
  var query MsgSlice
  for index, _ := range i {
    index_to_query := i[index] + uint64(index)*batch_sz
    cs, q := pi.Query(index_to_query, shared_state, p, DB.info) // TODO: Make batched queries smaller
    client_state = append(client_state, cs)
    query.data = append(query.data, q)
  }
  runtime.GC()
  printTime(start)
  fmt.Printf("\t\tOnline upload: %f KB\n",
             float64(query.size() * uint64(p.logq)) / float64(8 * 1024))
  bw += float64(query.size() * uint64(p.logq) / (8.0 * 1024.0))
  runtime.GC()

  fmt.Println("Answering query...")
  start = time.Now()
  answer := pi.Answer(DB, query, server_state, shared_state, p)
  elapsed := printTime(start)
  rate := printRate(p, elapsed, len(i))
  fmt.Printf("\t\tOnline download: %f KB\n",
             float64(answer.size() * uint64(p.logq)) / float64(8 * 1024))
  bw += float64(answer.size() * uint64(p.logq) / (8.0 * 1024.0))
  runtime.GC()

  pi.Reset(DB, p);
  fmt.Println("Reconstructing...")
  start = time.Now()

  for index, _ := range i {
    index_to_query := i[index] + uint64(index)*batch_sz
    val := pi.Recover(index_to_query, uint64(index), offline_download, answer,
                      client_state[index], p, DB.info)

    if (DB.GetElem(index_to_query) != val) {
      fmt.Printf("Batch %d (querying index %d -- row should be >= %d): Got %d instead of %d\n",
                 index, index_to_query, DB.data.rows/4, val, DB.GetElem(index_to_query))
      panic("Reconstruct failed!")
    }
  }
  fmt.Println("Success!")
  printTime(start)

  runtime.GC()
  debug.SetGCPercent(100)
  return rate, bw
}

// Test that DB packing methods are correct, when each database entry is ~ 1 Z_p elem.
func TestDBMediumEntries(t *testing.T) {
  N := uint64(4)
  d := uint64(9)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  vals := []uint64{1, 2, 3, 4}
  DB := MakeDB(N, d, &p, vals)
  if DB.info.packing != 1 || DB.info.ne != 1 {
    panic("Should not happen.")
  }

  for i := uint64(0); i < N; i++ {
    if DB.GetElem(i) != (i+1) {
      panic("Failure")
    }
  }
}

// Test that DB packing methods are correct, when multiple database entries fit in 1 Z_p elem.
func TestDBSmallEntries(t *testing.T) {
  N := uint64(4)
  d := uint64(3)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  vals := []uint64{1, 2, 3, 4}
  DB := MakeDB(N, d, &p, vals)
  if DB.info.packing <= 1 || DB.info.ne != 1 {
    panic("Should not happen.")
  }

  for i := uint64(0); i < N; i++ {
    if DB.GetElem(i) != (i+1) {
      panic("Failure")
    }
  }
}

// Test that DB packing methods are correct, when each database entry requires multiple Z_p elems.
func TestDBLargeEntries(t *testing.T) {
  N := uint64(4)
  d := uint64(12)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  vals := []uint64{1, 2, 3, 4}
  DB := MakeDB(N, d, &p, vals)
  if DB.info.packing != 0 || DB.info.ne <= 1 {
    panic("Should not happen.")
  }

  for i := uint64(0); i < N; i++ {
    if DB.GetElem(i) != (i+1) {
      panic("Failure")
    }
  }
}

// Print the BW used by SimplePIR
func TestSimplePirBW(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(2048)

  log_N, _ := strconv.Atoi(os.Getenv("LOG_N"))
  D, _ := strconv.Atoi(os.Getenv("D"))
  if log_N != 0 {
    N = uint64(1 << log_N)
  }
  if D != 0 {
    d = uint64(D)
  }

  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)
  DB := SetupDB(N, d, &p)

  fmt.Printf("Executing with entries consisting of %d (>= 1) bits; p is %d; packing factor is %d; number of DB elems per entry is %d.\n", 
             d, p.p, DB.info.packing, DB.info.ne)

  pir.GetBW(DB.info, p)
}

// Print the BW used by DoublePIR
func TestDoublePirBW(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(2048)

  log_N, _ := strconv.Atoi(os.Getenv("LOG_N"))
  D, _ := strconv.Atoi(os.Getenv("D"))
  if log_N != 0 {
    N = uint64(1 << log_N)
  }
  if D != 0 {
    d = uint64(D)
  }

  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)
  DB := SetupDB(N, d, &p)

  fmt.Printf("Executing with entries consisting of %d (>= 1) bits; p is %d; packing factor is %d; number of DB elems per entry is %d.\n", 
             d, p.p, DB.info.packing, DB.info.ne)

  pir.GetBW(DB.info, p)
}

// Test SimplePIR correctness on DB with short entries.
func TestSimplePir(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(8)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{262144})
}

// Test SimplePIR correctness on DB with long entries
func TestSimplePirLongRow(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(32)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{1})
}

// Test SimplePIR correctness on big DB
func TestSimplePirBigDB(t *testing.T) {
  N := uint64(1 << 25)
  d := uint64(7)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{0})
}

// Test SimplePIR correctness on DB with short entries, and batching.
func TestSimplePirBatch(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(8)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{0, 0, 0, 0})
}

// Test SimplePIR correctness on DB with long entries, and batching.
func TestSimplePirLongRowBatch(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(32)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{0, 0, 0, 0})
}

// Test DoublePIR correctness on DB with short entries.
func TestDoublePir(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(3)
  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{0})
}

// Test DoublePIR correctness on DB with long entries.
func TestDoublePirLongRow(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(32)
  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)

  fmt.Printf("Executing with entries consisting of %d (>= 1) bits; p is %d; packing factor is %d; number of DB elems per entry is %d.\n",
           d, p.p, DB.info.packing, DB.info.ne)

  RunPIR(&pir, DB, p, []uint64{1 << 19})
}

// Test DoublePIR correctness on big DB
func TestDoublePirBigDB(t *testing.T) {
  N := uint64(1 << 25)
  d := uint64(7)
  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{0})
}

// Test DoublePIR correctness on DB with short entries, and batching.
func TestDoublePirBatch(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(8)
  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{0, 0, 0, 0})
}

// Test DoublePIR correctness on DB with long entries, and batching.
func TestDoublePirLongRowBatch(t *testing.T) {
  N := uint64(1 << 20)
  d := uint64(32)
  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  DB := MakeRandomDB(N, d, &p)
  RunPIR(&pir, DB, p, []uint64{0, 0, 0, 0})
}

// Benchmark SimplePIR performance.
func BenchmarkSimplePirSingle(b *testing.B) {
  f, err := os.Create("simple-cpu.out")
  if err != nil {
    panic("Error creating file")
  }

  N := uint64(1 << 20)
  d := uint64(2048)
  
  log_N, _ := strconv.Atoi(os.Getenv("LOG_N"))
  D, _ := strconv.Atoi(os.Getenv("D"))
  if log_N != 0 {
    N = uint64(1 << log_N)
  }
  if D != 0 {
    d = uint64(D)
  }

  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  i := uint64(0) // index to query
  if (i >= p.l * p.m) {
    panic("Index out of dimensions")
  }

  DB := MakeRandomDB(N, d, &p)
  var tputs []float64
  for j := 0; j < 5; j++ {
    tput, _, _, _ := RunFakePIR(&pir, DB, p, []uint64{i}, f, false)
    tputs = append(tputs, tput)
  }
  fmt.Printf("Avg SimplePIR tput, except for first run: %f MB/s\n", avg(tputs))
  fmt.Printf("Std dev of SimplePIR tput, except for first run: %f MB/s\n", stddev(tputs))
}

// Benchmark DoublePIR performance.
func BenchmarkDoublePirSingle(b *testing.B) {
  f, err := os.Create("double-cpu.out")
  if err != nil {
    panic("Error creating file")
  }

  N := uint64(1 << 20)
  d := uint64(2048)
  
  log_N, _ := strconv.Atoi(os.Getenv("LOG_N"))
  D, _ := strconv.Atoi(os.Getenv("D"))
  if log_N != 0 {
    N = uint64(1 << log_N)
  }
  if D != 0 {
    d = uint64(D)
  }

  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  i := uint64(0) // index to query
  if (i >= p.l * p.m) {
    panic("Index out of dimensions")
  }

  DB := MakeRandomDB(N, d, &p)
  var tputs []float64
  for j := 0; j < 5; j++ {
    tput, _, _, _ := RunFakePIR(&pir, DB, p, []uint64{i}, f, false)
    tputs = append(tputs, tput)
  }
  fmt.Printf("Avg DoublePIR tput, except for first run: %f MB/s\n", avg(tputs))
  fmt.Printf("Std dev of DoublePIR tput, except for first run: %f MB/s\n", stddev(tputs))
}
/*
// Benchmark SimplePIR performance with batching.
func BenchmarkSimplePirBatch(b *testing.B) {
  f, err := os.Create("simple-cpu-batch.out")
  if err != nil {
    panic("Error creating file")
  }

  N := uint64(1 << 30)
  d := uint64(1)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  i := uint64(0) // index to query
  if (i >= p.l * p.m) {
    panic("Index out of dimensions")
  }

  DB := MakeRandomDB(N, d, p)
  //for j := 0; j < 10; j++ {
    RunFakePIR(&pir, DB, p, []uint64{i, i, i, i, i, i, i, i}, f, true)
  //}
}

// Benchmark DoublePIR performance with batching.
func BenchmarkDoublePirBatch(b *testing.B) {
  f, err := os.Create("double-cpu-batch.out")
  if err != nil {
    panic("Error creating file")
  }

  N := uint64(1 << 30)
  d := uint64(1)
  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  i := uint64(0) // index to query
  if (i >= p.l * p.m) {
    panic("Index out of dimensions")
  }

  DB := MakeRandomDB(N, d, p)
  //for j := 0; j < 10; j++ {
    RunFakePIR(&pir, DB, p, []uint64{i, i, i, i, i, i, i, i}, f, true)
  //}
}*/

// Benchmark SimplePIR performance.
func BenchmarkSimplePirVaryingDB(b *testing.B) {
  flog, err := os.OpenFile("simple-comm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    panic("Error creating log file")
  }
  defer flog.Close()

  writer := csv.NewWriter(flog)
  defer writer.Flush()

  records := []string{"N", "d", "tput", "tput_stddev", "offline_comm", "online_comm"}
  writer.Write(records)

  pir := SimplePIR{}

  // Set N, D
  total_sz := 33

  for d := uint64(1); d <= 32768; d*= 2 {
    N := uint64(1 << total_sz) / d
    p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

    i := uint64(0) // index to query
    if (i >= p.l * p.m) {
      panic("Index out of dimensions")
    }

    DB := MakeRandomDB(N, d, &p)
    var tputs []float64
    var offline_cs []float64
    var online_cs []float64
    
    for j := 0; j < 1; j++ {
      tput, _, offline_c, online_c := RunFakePIR(&pir, DB, p, []uint64{i}, nil, false)
      tputs = append(tputs, tput)
      offline_cs = append(offline_cs, offline_c)
      online_cs = append(online_cs, online_c)
    }
    fmt.Printf("Avg SimplePIR tput (%d, %d), except for first run: %f MB/s\n", N, d, avg(tputs))
    fmt.Printf("Std dev of SimplePIR tput (%d, %d), except for first run: %f MB/s\n", N, d, stddev(tputs))
    if (stddev(offline_cs) != 0) || (stddev(online_cs) != 0) {
      fmt.Printf("%f %f SHOULD NOT HAPPEN\n", stddev(offline_cs), stddev(online_cs))
      //panic("Should not happen!")
    }
    writer.Write([]string{strconv.FormatUint(N, 10), 
                          strconv.FormatUint(d, 10),
			  strconv.FormatFloat(avg(tputs), 'f', 4, 64), 
			  strconv.FormatFloat(stddev(tputs), 'f', 4, 64), 
			  strconv.FormatFloat(avg(offline_cs), 'f', 4, 64),
			  strconv.FormatFloat(avg(online_cs), 'f', 4, 64)})
  }
}

// Benchmark DoublePIR performance.
func BenchmarkDoublePirVaryingDB(b *testing.B) {
  flog, err := os.OpenFile("double-comm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
    panic("Error creating log file")
  }
  defer flog.Close()

  writer := csv.NewWriter(flog)
  defer writer.Flush()

  records := []string{"N", "d", "tput", "tput_stddev", "offline_comm", "online_comm"}
  writer.Write(records)

  pir := DoublePIR{}

  // Set N, D
  total_sz := 33

  for d := uint64(1); d <= 32768; d*= 2 {
    N := uint64(1 << total_sz) / d
    p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

    i := uint64(0) // index to query
    if (i >= p.l * p.m) {
      panic("Index out of dimensions")
    }

    DB := MakeRandomDB(N, d, &p)
    var tputs []float64
    var offline_cs []float64
    var online_cs []float64

    for j := 0; j < 1; j++ {
      tput, _, offline_c, online_c := RunFakePIR(&pir, DB, p, []uint64{i}, nil, false)
      tputs = append(tputs, tput)
      offline_cs = append(offline_cs, offline_c)
      online_cs = append(online_cs, online_c)
    }
    fmt.Printf("Avg SimplePIR tput (%d, %d), except for first run: %f MB/s\n", N, d, avg(tputs))
    fmt.Printf("Std dev of SimplePIR tput (%d, %d), except for first run: %f MB/s\n", N, d, stddev(tputs))
    if (stddev(offline_cs) != 0) || (stddev(online_cs) != 0) {
      fmt.Printf("%f %f SHOULD NOT HAPPEN\n", stddev(offline_cs), stddev(online_cs))
      //panic("Should not happen!")
    }
    writer.Write([]string{strconv.FormatUint(N, 10),
                          strconv.FormatUint(d, 10),
                          strconv.FormatFloat(avg(tputs), 'f', 4, 64),
                          strconv.FormatFloat(stddev(tputs), 'f', 4, 64),
                          strconv.FormatFloat(avg(offline_cs), 'f', 4, 64),
                          strconv.FormatFloat(avg(online_cs), 'f', 4, 64)})
  }
}

// Benchmark SimplePIR performance with (large) batching.
func BenchmarkSimplePirBatchLarge(b *testing.B) {
  f, err := os.Create("simple-cpu-batch.out")
  if err != nil {
    panic("Error creating file")
  }

  flog, err := os.Create("simple-batch.log")
  if err != nil {
    panic("Error creating log file")
  }
  defer flog.Close()

  N := uint64(1 << 33)
  d := uint64(1)
  pir := SimplePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  i := uint64(0) // index to query
  if (i >= p.l * p.m) {
    panic("Index out of dimensions")
  }

  DB := MakeRandomDB(N, d, &p)

  writer := csv.NewWriter(flog)
  defer writer.Flush()

  records := []string{"Batch_sz", "Good_tput", "Good_std_dev", "Num_successful_queries", "Tput"}
  writer.Write(records)

  for trial := 0; trial <= 10; trial += 1 {
    batch_sz := (1 << trial)
    var query []uint64
    for j := 0; j < batch_sz; j++ {
      query = append(query, i)
    }
    var tputs []float64
    for iter := 0; iter < 5; iter++ {
      tput, _, _, _ := RunFakePIR(&pir, DB, p, query, f, false)
      tputs = append(tputs, tput)
    }

    expected_num_empty_buckets := math.Pow(float64(batch_sz-1)/float64(batch_sz), float64(batch_sz)) * float64(batch_sz)
    expected_num_successful_queries := float64(batch_sz) - expected_num_empty_buckets
    good_tput := avg(tputs) / float64(batch_sz) * expected_num_successful_queries
    dev := stddev(tputs) / float64(batch_sz) * expected_num_successful_queries

    writer.Write([]string{strconv.Itoa(batch_sz), 
                          strconv.FormatFloat(good_tput, 'f', 4, 64), 
                          strconv.FormatFloat(dev, 'f', 4, 64), 
			  strconv.FormatFloat(expected_num_successful_queries, 'f', 4, 64), 
			  strconv.FormatFloat(avg(tputs), 'f', 4, 64)})
  }
}

// Benchmark DoublePIR performance with (large) batching.
func BenchmarkDoublePirBatchLarge(b *testing.B) {
  f, err := os.Create("double-cpu-batch.out")
  if err != nil {
    panic("Error creating file")
  }

  flog, err := os.Create("double-batch.log")
  if err != nil {
    panic("Error creating log file")
  }
  defer flog.Close()

  N := uint64(1 << 33)
  d := uint64(1)
  pir := DoublePIR{}
  p := pir.PickParams(N, d, SEC_PARAM, LOGQ)

  i := uint64(0) // index to query
  if (i >= p.l * p.m) {
    panic("Index out of dimensions")
  }

  DB := MakeRandomDB(N, d, &p)

  writer := csv.NewWriter(flog)
  defer writer.Flush()

  records := []string{"Batch_sz", "Good_tput", "Good_std_dev", "Num_successful_queries", "Tput"}
  writer.Write(records)

  for trial := 0; trial <= 10; trial += 1 {
    batch_sz := (1 << trial)
    var query []uint64
    for j := 0; j < batch_sz; j++ {
      query = append(query, i)
    }
    var tputs []float64
    for iter := 0; iter < 5; iter++ {
      tput, _, _, _ := RunFakePIR(&pir, DB, p, query, f, false)
      tputs = append(tputs, tput)
    }
    expected_num_empty_buckets := math.Pow(float64(batch_sz-1)/float64(batch_sz), float64(batch_sz)) * float64(batch_sz)
    expected_num_successful_queries := float64(batch_sz) - expected_num_empty_buckets
    good_tput := avg(tputs) / float64(batch_sz) * expected_num_successful_queries
    dev := stddev(tputs) / float64(batch_sz) * expected_num_successful_queries

    writer.Write([]string{strconv.Itoa(batch_sz), 
                          strconv.FormatFloat(good_tput, 'f', 4, 64), 
                          strconv.FormatFloat(dev, 'f', 4, 64), 
			  strconv.FormatFloat(expected_num_successful_queries, 'f', 4, 64), 
			  strconv.FormatFloat(avg(tputs), 'f', 4, 64)})
  }
}
