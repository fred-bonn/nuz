package engine

import "testing"

func newEmptyActionQueue() *actionQueue {
	return &actionQueue{
		queue: make(priorityQueue[action], 0, 5),
	}
}

// BenchmarkActionQueueInit measures the cost of initializing the queue for a turn:
// creating a new queue, pushing actions, and then sorting for execution.
func BenchmarkActionQueueInit(b *testing.B) {
	actions := []action{
		&dummyAction{priority: 1, spd: 50},
		&dummyAction{priority: 5, spd: 30},
		&dummyAction{priority: 3, spd: 70},
		&dummyAction{priority: 2, spd: 40},
		&dummyAction{priority: 4, spd: 60},
	}
	bs := &dummyBattleState{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := newEmptyActionQueue()
		bs.actions = q
		for _, a := range actions {
			q.queue.push(a)
		}
		q.sort(bs)
	}
}

// BenchmarkActionQueueDrain measures the cost of draining the queue after sorting.
// Reuses the same queue allocation across iterations to focus on the drain cost.
func BenchmarkActionQueueDrain(b *testing.B) {
	actions := []action{
		&dummyAction{priority: 2, spd: 50},
		&dummyAction{priority: 5, spd: 30},
		&dummyAction{priority: 3, spd: 70},
		&dummyAction{priority: 2, spd: 40},
		&dummyAction{priority: 4, spd: 60},
	}
	bs := &dummyBattleState{}
	q := newEmptyActionQueue()
	bs.actions = q

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Push actions to the queue
		for _, a := range actions {
			q.queue.push(a)
		}
		q.sort(bs)
		// Drain the queue
		for len(q.queue) > 0 {
			q.queue.pop()
		}
	}
}
