package util

import "container/heap"

type internalIntHeap []uint64

func (h internalIntHeap) Len() int           { return len(h) }
func (h internalIntHeap) Less(i, j int) bool { return h[j] < h[i] }
func (h internalIntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *internalIntHeap) Push(x any) {
	*h = append(*h, x.(uint64))
}
func (h *internalIntHeap) Pop() any {
	old := *h
	size := len(old)
	result := old[size-1]
	*h = old[0 : size-1]
	return result
}

type LowestNIntHeap struct {
	heap  internalIntHeap
	limit int
}

func LowestNIntHeap_For(limit int) LowestNIntHeap {
	collect := LowestNIntHeap{limit: limit}
	heap.Init(&collect.heap)
	return collect
}

func (collect *LowestNIntHeap) Offer(value uint64) bool {
	// DOCO: The minimum element in the tree is the root, at index 0.
	// But we're using reversed order and thus its the highest/worst for us
	if collect.heap.Len() < collect.limit || collect.heap.Len() == 0 {
		heap.Push(&collect.heap, value)
		return true
	} else if value < collect.heap[0] {
		collect.heap[0] = value
		heap.Fix(&collect.heap, 0)
		return true
	}
	return false
}
