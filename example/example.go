// Copyright © 2014-2019 Lawrence E. Bakst. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/trace"
	"sync"
	"time"

	_ "leb.io/hrff"
	"leb.io/runstats"
)

var gcp = flag.Int("gcp", 100, "gc triggers when ratio of new data to live data reaches this percentage")
var mstats = flag.Bool("m", true, "print memory stats")
var ostats = flag.Bool("o", false, "print system overhead stats")
var astats = flag.Bool("a", false, "print detaied allocation size stats")
var gcstats = flag.Bool("g", false, "print detaied gc pause stats and gc pause time stats")
var ngc = flag.Int("ngc", 8, "print only the last ngc entries from gc ciculrar buffers")
var allocs = flag.Int("allocs", 500, "number of allocations to make")
var iterations = flag.Int("iterations", 5, "number of allocation interations")
var delay = flag.Duration("delay", 0*time.Second, "delay between interations")
var ngr = flag.Int("ngr", 1, "number of go routines")
var dur = flag.Duration("d", 1*time.Second, "duration for stats")
var trc = flag.Bool("trace", true, "run a trace")
var tf = flag.String("tf", "trace.out", "name of trace file")
var ps = flag.Bool("ps", false, "print stats")
var lt = flag.Bool("lt", false, "lock go routine to thread")

var s = rand.NewSource(time.Now().UTC().UnixNano())
var r = rand.New(s)

// rbetween returns random int [a, b]
func rbetween(r *rand.Rand, a int, b int) int {
	return r.Intn(b-a+1) + a
}

func alloc(wg *sync.WaitGroup, inst int) {
	var bss = []int{1024, 2048, 4096, 8192, 1 * 1024 * 1024}
	var bsr [][]byte

	if *lt {
		runtime.LockOSThread()
	}
	for i := 0; i < *iterations; i++ {
		for j := 0; j < *allocs; j++ {
			siz := rbetween(r, 0, len(bss)-1)
			b := make([]byte, bss[siz], bss[siz])
			bsr = append(bsr, b)
		}
		if *delay != 0*time.Second {
			time.Sleep(*delay)
		}
	}
	wg.Done()
}

func printStats() {
	var m runtime.MemStats

	ticker := time.NewTicker(*dur)
	for {
		<-ticker.C
		runtime.ReadMemStats(&m)
		runstats.PrintMemStats(&m, *mstats, *ostats, *astats, *gcstats, *ngc)
		fmt.Printf("\n")
	}
}

func main() {
	var wg sync.WaitGroup
	var f *os.File
	var err error

	flag.Parse()
	debug.SetGCPercent(*gcp)
	numCPU := runtime.NumCPU()
	maxProcs := runtime.GOMAXPROCS(0)
	fmt.Printf("start: numCPU=%d, maxProcs=%d\n", numCPU, maxProcs)
	if *trc {
		f, err = os.Create(*tf) // os.OpenFile("trace.out", os.O_CREATE, 0666)
		if f == nil || err != nil {
			fmt.Printf("reader: |%s| could not be opened\n", "")
			return
		}
		defer f.Close()
	}
	if *trc {
		trace.Start(f)
		defer trace.Stop()
	}
	wg.Add(*ngr)
	for i := 0; i < *ngr; i++ {
		//fmt.Printf("start %d\n", i)
		go alloc(&wg, i)
	}
	if *ps {
		go printStats()
	}
	wg.Wait()
	fmt.Printf("end\n")
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
