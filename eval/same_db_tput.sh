# !/bin/bash/

cd ../../spiral/
python3 /home/ubuntu/spiral/run_all.py table > final_run_same_db_rel

cd ../latpir/go
LOG_N=33 D=1 go test -bench PirSingle -timeout 0 -run=^$ > final_run_same_db_this
