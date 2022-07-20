#!/bin/bash

go test -bench Pir -timeout 0 -run=^$
go test -bench Xor -timeout 0 -run=^$
