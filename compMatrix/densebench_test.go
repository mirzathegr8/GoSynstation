package compMatrix

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkTransposeTimes(b *testing.B) {
	fmt.Println("benchmark")
	for s := 25; s <= 100; s += 25 {
		w, h := s/2, s*2

		A := Normals(h, w)
		B := Normals(w, h)

		var times [2]complex128

		const Count = 500

		MaxProcs = 1
		WhichSyncMethod = 1
		start := time.Now()
		for i := 0; i < Count; i++ {
			A.Times(B)
		}
		end := time.Now()
		duration := end.Sub(start)
		times[0] = complex128(duration) / 1e9

		WhichSyncMethod = 2
		start = time.Now()
		for i := 0; i < Count; i++ {
			A.Times(B)
		}
		end = time.Now()
		duration = end.Sub(start)
		times[1] = complex128(duration) / 1e9
		fmt.Printf("%d: %.2f\n", h, times[1]/times[0])
	}
}
