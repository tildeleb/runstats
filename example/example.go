// Copyright © 2014-2019 Lawrence E. Bakst. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	_ "leb.io/hrff"
	"leb.io/runstats"
)

var gcp = flag.Int("gcp", 100, "gc triggers when ratio of new data to live data reaches this percentage")
var mstats = flag.Bool("m", true, "print memory stats")
var ostats = flag.Bool("o", false, "print system overhead stats")
var astats = flag.Bool("a", true, "print detaied allocation size stats")
var gcstats = flag.Bool("g", true, "print detaied gc pause stats and gc pause time stats")
var ngc = flag.Int("ngc", 8, "print only the last ngc entries from gc ciculrar buffers")

var s = rand.NewSource(time.Now().UTC().UnixNano())
var r = rand.New(s)

// rbetween returns random int [a, b]
func rbetween(r *rand.Rand, a int, b int) int {
	return r.Intn(b-a+1) + a
}

func alloc() {
	var m runtime.MemStats
	var bss = []int{1024, 2048, 4096, 8192, 1 * 1024 * 1024}
	var bsr [][]byte
	var n = 1500
	const slp = 1.0
	for {
		for i := 0; i < n; i++ {
			siz := rbetween(r, 0, len(bss)-1)
			b := make([]byte, bss[siz], bss[siz])
			bsr = append(bsr, b)
		}
		runtime.ReadMemStats(&m)
		runstats.PrintMemStats(&m, *mstats, *ostats, *astats, *gcstats, *ngc)
		fmt.Printf("\n")
		time.Sleep(slp * time.Second)
	}
}

/*
gctrace: setting gctrace=1 causes the garbage collector to emit a single line to standard
error at each collection, summarizing the amount of memory collected and the
length of the pause. The format of this line is subject to change.
Currently, it is:
	gc # @#s #%: #+#+# ms clock, #+#/#/#+# ms cpu, #->#-># MB, # MB goal, # P
where the fields are as follows:
	gc #        the GC number, incremented at each GC
	@#s         time in seconds since program start
	#%          percentage of time spent in GC since program start
	#+...+#     wall-clock/CPU times for the phases of the GC
	#->#-># MB  heap size at GC start, at GC end, and live heap
	# MB goal   goal heap size
	# P         number of processors used
The phases are stop-the-world (STW) sweep termination, concurrent
mark and scan, and STW mark termination. The CPU times
for mark/scan are broken down in to assist time (GC performed in
line with allocation), background GC time, and idle GC time.
If the line ends with "(forced)", this GC was forced by a
runtime.GC() call.
*/

func main() {
	var wg sync.WaitGroup

	flag.Parse()
	debug.SetGCPercent(*gcp)
	numCPU := runtime.NumCPU()
	maxProcs := runtime.GOMAXPROCS(0)
	fmt.Printf("numCPU=%d, maxProcs=%d\n", numCPU, maxProcs)
	wg.Add(1)
	alloc()
	wg.Wait()
}
