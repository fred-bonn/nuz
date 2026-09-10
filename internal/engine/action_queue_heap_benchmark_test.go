package engine

import (
	"container/heap"
	"testing"
)

func newEmptyActionQueueHeap() *ActionQueueOld {
	res := make(ActionQueueOld, 0, 5)
	return &res
}

// BenchmarkActionQueueHeapInit measures the cost of initializing the heap:
// creating a new heap, initializing it, and pushing actions for execution.
func BenchmarkActionQueueHeapInit(b *testing.B) {
	actions := []action{
		&dummyAction{priority: 1, spd: 50},
		&dummyAction{priority: 5, spd: 30},
		&dummyAction{priority: 3, spd: 70},
		&dummyAction{priority: 2, spd: 40},
		&dummyAction{priority: 4, spd: 60},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := newEmptyActionQueueHeap()
		heap.Init(q)
		for _, a := range actions {
			heap.Push(q, a)
		}
	}
}

// BenchmarkActionQueueHeapDrain measures the cost of draining the heap after initialization.
// Reuses the same heap allocation across iterations to focus on the drain cost.
func BenchmarkActionQueueHeapDrain(b *testing.B) {
	actions := []action{
		&dummyAction{priority: 2, spd: 50},
		&dummyAction{priority: 5, spd: 30},
		&dummyAction{priority: 3, spd: 70},
		&dummyAction{priority: 2, spd: 40},
		&dummyAction{priority: 4, spd: 60},
	}
	q := newEmptyActionQueueHeap()
	heap.Init(q)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Push actions to the heap
		for _, a := range actions {
			heap.Push(q, a)
		}
		// Drain the heap
		for q.Len() > 0 {
			_ = heap.Pop(q).(action)
		}
	}
}
