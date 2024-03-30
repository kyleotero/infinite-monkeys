package benchmark

import (
	"math"
	"sync/atomic"
)

type Benchmark struct {
	last  int64
	sum   int64
	count int32
}

func (b *Benchmark) LogTimestamp(timestamp int64) {
	prevLast := atomic.SwapInt64(&b.last, timestamp)
	if prevLast != 0 {
		atomic.AddInt64(&b.sum, timestamp-prevLast)
	}
	atomic.AddInt32(&b.count, 1)
}

func (b *Benchmark) OutputAvg() float64 {
	if c := atomic.LoadInt32(&b.count); c > 0 {
		return float64(atomic.LoadInt64(&b.sum)) / float64(c)
	}
	return math.NaN()
}
