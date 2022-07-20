# Simple and Fast Single-Server Private Information Retrieval

This directory contains the code for SimplePIR and DoublePIR, two high-throughput single-server PIR schemes presented in the paper "Simple and Fast Single-Server Private Information Retrieval" by Alexandra Henzinger, Matthew M. Hong, Henry Corrigan-Gibbs, Sarah Meiklejohn, and Vinod Vaikuntanathan.

**Warning**: This code is a research prototype.

## Overview

The `pir/` directory contains the code for SimplePIR and DoublePIR. In particular, it contains the files:
- `pir.go`, which defines the interface for a PIR with preprocessing scheme, and `simple_pir.go` and `double_pir.go`, which implement this interface.
- `pir_test.go`, which contains correctness tests and performance benchmarks for the SimplePIR and DoublePIR implementations. 
- `pir.h` and `pir.c`, which implement matrix multiplication routines.
- `matrix.go`, which implements other operations on matrices.
- `database.go`, which implements operations on databases to transform them to the format used by SimplePIR and DoublePIR.
- `params.csv`, which contains the LWE parameters used in this work, and `params.go`, which selects the appropriate LWE parameters based on the PIR database dimensions.
- `gauss.go`, which implements sampling from a discrete Guassian distribution with a fixed, hard-coded variance.
- `rand.go`, `logging.go`, and `utils.go`, which implement helper routines to sample cryptographic randomness and log debugging information, among others.

The `eval/` directory contains scripts to generate Figure 8 from the paper.

It is worth noting that our performance benchmarks run on random databases and skip the preprocessing step (i.e., they use randomly generated hints), so that these tests take less time to execute. On the other hand, our correctness tests run on random databases, perform the full preprocessing step, and check that the PIR outputs are correct.  

## Setup

To run the SimplePIR and DoublePIR code, install [Go 1.18](https://go.dev/).

To produce the plots, additionally install [Python 3](https://www.python.org/downloads/) and [Matplotlib](https://matplotlib.org/).

## Usage

To run all SimplePIR and DoublePIR tests (including the correctness tests), run 
```
cd pir/
go test
cd ..
``` 

To analytically compute SimplePIR and DoublePIR's communication on a database of $2^n$ entries, each consisting of $d$ bits, run 
```
cd pir/
LOG_N=n D=d go test -run=BW
cd ..
```

To benchmark SimplePIR and DoublePIR's performance on a database of $2^n$ entries, each consisting of $d$ bits, run 
```
cd pir/
LOG_N=n D=d go test -bench PirSingle -timeout 0 -run=^$
cd ..
``` 
(SimplePIR and DoublePIR's maximal throughput is achieved with $n = 22$ and $d = 2048$.)

To benchmark SimplePIR and DoublePIR's performance on a database of $2^n$ entries, each consisting of $d$ bits, with batches of increasing size, run 
```
cd pir/
LOG_N=n D=d go test -bench PirBatch -timeout 0 -run=^$
cd ..
``` 

To produce a plot of SimplePIR and DoublePIR's throughput with increasing batch sizes, first run the command above to benchmark the scheme's performance on a database of the desired size. Then, run
```
cd eval/
python3 plot.py -p batch_tput -f ../pir/simple-batch.log ../pir/double-batch.log -n SimplePIR DoublePIR
cd ..
```

## Citation

```
@misc{simplepir,
    author =  {Alexandra Henzinger and Matthew M. Hong and Henry Corrigan-Gibbs and Sarah Meiklejohn and Vinod Vaikuntanathan},
    title  =  {Simple and Fast Single-Server Private Information Retrieval},
    year   =  {2022},
}
```
