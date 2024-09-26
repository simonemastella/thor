package schedule

import (
	"container/heap"
	"sync"
	"time"

	"github.com/vechain/thor/v2/tx"
)

type Item struct {
	Tx   *tx.Transaction
	Date time.Time
}

type minHeap []*Item

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].Date.Before(h[j].Date) }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x interface{}) {
	item := x.(*Item)
	*h = append(*h, item)
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[0 : n-1]
	return item
}

type HeapManager struct {
	heap minHeap
	mu   sync.RWMutex
}

func NewHeapManager() *HeapManager {
	return &HeapManager{
		heap: make(minHeap, 0),
	}
}

func (hm *HeapManager) Push(item *Item) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	heap.Push(&hm.heap, item)
}

func (hm *HeapManager) Pop() *Item {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	if hm.heap.Len() > 0 {
		return heap.Pop(&hm.heap).(*Item)
	}
	return nil
}

func (hm *HeapManager) Top() *Item {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	if hm.heap.Len() > 0 {
		return hm.heap[0]
	}
	return nil
}

func (hm *HeapManager) Len() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.heap.Len()
}

type Schedule struct {
	heapManager *HeapManager
}

func NewSchedule() *Schedule {
	return &Schedule{
		heapManager: NewHeapManager(),
	}
}

func (s *Schedule) Push(tx *tx.Transaction, date time.Time) {
	s.heapManager.Push(&Item{Tx: tx, Date: date})
}

func (s *Schedule) Pop() *Item {
	return s.heapManager.Pop()
}

func (s *Schedule) Top() *Item {
	return s.heapManager.Top()
}

func (s *Schedule) Len() int {
	return s.heapManager.Len()
}
