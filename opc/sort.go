package opc

import (
	"iter"
	"slices"
)

// sortedKV iterates over m in key-sorted order. Deterministic output is
// important so that re-serializing a parsed package produces identical bytes
// (modulo ZIP timestamps).
func sortedKV[K ~string, V any](m map[K]V) iter.Seq2[K, V] {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return func(yield func(K, V) bool) {
		for _, k := range keys {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}
