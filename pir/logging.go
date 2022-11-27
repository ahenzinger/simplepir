package pir

import "time"
import "fmt"
import "os"
import "bufio"
import "math"

func printTime(start time.Time) time.Duration {
	elapsed := time.Since(start)
	fmt.Printf("\tElapsed: %s\n", elapsed)
	return elapsed
}

func printRate(p Params, elapsed time.Duration, batch_sz int) float64 {
	rate := math.Log2(float64((p.P))) * float64(p.L*p.M) * float64(batch_sz) /
		float64(8*1024*1024*elapsed.Seconds())
	fmt.Printf("\tRate: %f MB/s\n", rate)
	return rate
}

func clearFile(filename string) {
	f, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err = f.WriteString("log(n) log(l) log(m) log(q) rate(MB/s) BW(KB)\n"); err != nil {
		panic(err)
	}
}

func writeToFile(p Params, rate, bw float64, filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintf(w,
		"%d,%d,%d,%d,%f,%f\n",
		int(math.Log2(float64(p.N))),
		int(math.Log2(float64(p.L))),
		int(math.Log2(float64(p.M))),
		p.Logq,
		rate,
		bw)
	w.Flush()
}
