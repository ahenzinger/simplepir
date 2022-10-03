#!/bin/bash

go test -bench PirVaryingDB -timeout 0 -run=^$ | tee results/our_pir_varying_db.txt
LOG_N=33 D=1 go test -bench PirSingle -timeout 0 -run=^$ | tee results/our_pir_same_db_tput.txt
go test -bench PirBatchLarge -timeout 0 -run=^$ | tee results/our_pir_batch.txt
LOG_N=36 D=1 go test -bench PirSingle -timeout 0 -run=^$ | tee results/our_pir_ct_app.txt
