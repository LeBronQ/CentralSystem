package main

import (
	"testing"
)

func Benchmark_main(b *testing.B) {
	for n := 0; n < b.N; n++ {
		main()
	}
}
