package main

import "testing"

// go test -bench=. -test.benchmem=true -test.benchtime=20s

func BenchmarkDoRequestReadAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DoRequestReadAll()
	}
}

func BenchmarkDoRequestDiscardCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DoRequestDiscardCopy()
	}
}
