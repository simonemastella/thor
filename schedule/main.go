// Copyright (c) 2024 Simone Mastella (pussy smasher)

// Distributed under the GNU Lesser General Public License v3.0 software license, see the accompanying
// file LICENSE or <https://www.gnu.org/licenses/lgpl-3.0.html>

package main

import (
	"container/heap"
	"fmt"
	"time"
)

type Item struct {
	value string
	date  time.Time
	index int
}

type MinHeap []*Item

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].date.Before(h[j].date) }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }

func (h *MinHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*Item)
	item.index = n
	*h = append(*h, item)
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // evita memory leak
	item.index = -1 // per sicurezza
	*h = old[0 : n-1]
	return item
}

func main() {
	h := &MinHeap{}
	heap.Init(h)

	// Inserisci alcuni elementi
	heap.Push(h, &Item{value: "2", date: time.Now().Add(-2 * time.Hour)})
	heap.Push(h, &Item{value: "3", date: time.Now().Add(-1 * time.Hour)})
	heap.Push(h, &Item{value: "1", date: time.Now().Add(-4 * time.Hour)})
	heap.Push(h, &Item{value: "0", date: time.Now().Add(-8 * time.Hour)})
	heap.Push(h, &Item{value: "4", date: time.Now()})

	// Estrai gli elementi (verranno estratti in ordine di data)
	for h.Len() > 0 {
		item := heap.Pop(h).(*Item)
		fmt.Printf("Valore: %s, Data: %v\n", item.value, item.date)
	}
}
