// Copyright © 2014-2019 Lawrence E. Bakst. All rights reserved.

// Cleaved from cuckoo
// runstats provides some simple functions for printing memory and gc stats

package runstats

import (
	"fmt"
	"leb.io/hrff"
	"runtime"
	"time"
)

func hu(v uint64, u string) hrff.Int64 {
	return hrff.Int64{V: int64(v), U: u}
}

func hi(v int64, u string) hrff.Int64 {
	return hrff.Int64{V: int64(v), U: u}
}

// The pause buffer is circular. The most recent pause is at
// pause_ns[(numgc-1)%len(pause_ns)], and then backward
// from there to go back farther in time. We deliver the times
// most recent first (in p[0]).

// PrintCircularBuffer print a cicurlar buffer
func PrintCircularBuffer(label string, max int, n int, b bool, s [256]uint64) {
	if n > 0 {
		cnt := 0
		length := len(s)
		fmt.Printf("%s: ", label) // "PauseNs: "
		for i := range s {
			idx := (n - 1 - i) % length
			if idx < 0 {
				idx += 256
			}
			p := s[idx]
			if p != 0 {
				if i != 0 {
					fmt.Printf(", ")
				}
				cnt++
				if max > 0 && cnt >= max {
					break
				}
				if b {
					et := time.Duration(int64(p))
					fmt.Printf("%d: %v", n-i, et)
				} else {
					t := time.Unix(int64(0), int64(p))
					fmt.Printf("%d: %s", n-i, t.Format("15:04:05.99"))
				}
			}
		}
	}
	fmt.Printf("\n")
}

// PrintMemStats prints mem and gc stats
func PrintMemStats(m *runtime.MemStats, mstats, ostats, astats, gc bool, pauses int) {
	if mstats {
		fmt.Printf("Alloc=%h, TotalAlloc=%h, Sys=%h, Lookups=%h, Mallocs=%h, Frees=%h\n",
			hu(m.Alloc, "B"), hu(m.TotalAlloc, "B"), hu(m.Sys, "B"), hu(m.Lookups, ""), hu(m.Mallocs, ""), hu(m.Frees, ""))
		fmt.Printf("HeapAlloc=%h, HeapSys=%h, HeapIdle=%h, HeapInuse=%h, HeapReleased=%h, HeapObjects=%h, StackInuse=%h, StackSys=%h\n",
			hu(m.HeapAlloc, "B"), hu(m.HeapSys, "B"), hu(m.HeapIdle, "B"), hu(m.HeapInuse, "B"), hu(m.HeapReleased, "B"),
			hu(m.HeapObjects, ""), hu(m.StackInuse, "B"), hu(m.StackSys, "B"))
		if ostats {
			fmt.Printf("MSpanInuse=%d, MSpanSys=%d, m.MCacheInuse=%d, MCacheSys=%d, BuckHashSys=%d, GCSys=%d, OtherSys=%d\n",
				m.MSpanInuse, m.MSpanSys, m.MCacheInuse, m.MCacheSys, m.BuckHashSys, m.GCSys, m.OtherSys)
		}

		t1 := time.Unix(0, int64(m.LastGC))
		//t2 := time.Now()
		//t3 := time.Unix(int64(0), int64(m.PauseTotalNs))
		et := time.Duration(int64(m.PauseTotalNs)) // Since(t3)
		fmt.Printf("NextGC=%h, NumGC=%d, LastGC=%s, PauseTotalNs=%v, NumForcedGC=%d, GCCPUFraction=%0.2f\n",
			hu(m.NextGC, "B"), m.NumGC, t1.Format("15:04:05.99"), et, m.NumForcedGC, m.GCCPUFraction)
	}
	fmt.Printf("\n")

	if astats {
		for i, b := range m.BySize {
			if b.Mallocs == 0 {
				continue
			}
			fmt.Printf("BySize[%d]: Size=%d, Malloc=%d, Frees=%d\n", i, b.Size, b.Mallocs, b.Frees)
		}
		fmt.Printf("\n")
	}

	if gc {
		PrintCircularBuffer("PauseNs", pauses, int(m.NumGC), true, m.PauseNs)
		PrintCircularBuffer("PauseEnd", pauses, int(m.NumGC), false, m.PauseEnd)
		fmt.Printf("\n")
	}
}
