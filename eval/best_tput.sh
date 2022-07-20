#python3 /home/ubuntu/spiral/run_scheme.py sealpir 22 288 --show-output 
#python3 /home/ubuntu/spiral/run_scheme.py fastpir 20 1024 --show-output 
#python3 /home/ubuntu/spiral/run_scheme.py onionpir 15 30720 --show-output 
python3 /home/ubuntu/spiral/run_scheme.py spiral 14 102400 --show-output 
python3 /home/ubuntu/spiral/run_scheme.py spiral-pack 15 30720 --show-output 
python3 /home/ubuntu/spiral/run_scheme.py spiralstream 15 30720 --show-output 
python3 /home/ubuntu/spiral/run_scheme.py spiralstream-pack 15 30720 --show-output 
#python3 /home/ubuntu/spiral/run_scheme.py dpfpir 25 32 --show-output 


#LOG_N=22 D=2048 go test -bench PirSingle -timeout 0 -run=^
#go test -bench Xor -timeout 0 -run=^$
