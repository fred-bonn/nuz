package engine

import "testing"

func TestPriorityQueuePushesElementsInOrder(t *testing.T) {
	tests := map[string]struct {
		queue priorityQueue[int]
		push  []int
		want  priorityQueue[int]
	}{
		"pushes into an empty queue":                      {queue: priorityQueue[int]{}, push: []int{1, 2, 3}, want: priorityQueue[int]{1, 2, 3}},
		"pushes multiple elements into a populated queue": {queue: priorityQueue[int]{1, 2, 3}, push: []int{1, 2, 3}, want: priorityQueue[int]{1, 2, 3, 1, 2, 3}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, element := range tc.push {
				tc.queue.push(element)
			}
			if n := len(tc.queue); n != len(tc.want) {
				t.Fatalf("%s: tc.queue.len() = %d, want %d", name, n, len(tc.want))
			}
			for i, r := range tc.want {
				if r != tc.queue[i] {
					t.Errorf("%s: mismatch at index %d, %d != %d", name, i, r, tc.queue[i])
				}
			}
		})
	}
}

func TestPriorityQueuePopsElementsInPriorityOrder(t *testing.T) {
	tests := map[string]struct {
		queue    priorityQueue[int]
		want     int
		wantOk   bool
		wantLeft priorityQueue[int]
	}{
		"pops from an empty queue":                              {queue: priorityQueue[int]{}, want: 0, wantOk: false, wantLeft: priorityQueue[int]{}},
		"pops the highest priority element from a larger queue": {queue: priorityQueue[int]{1, 2, 3}, want: 3, wantOk: true, wantLeft: priorityQueue[int]{1, 2}},
		"pops the only element from a single-item queue":        {queue: priorityQueue[int]{1}, want: 1, wantOk: true, wantLeft: priorityQueue[int]{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res, ok := tc.queue.pop()
			if tc.wantOk && !ok {
				t.Fatalf("%s: queue is empty when it wasn't expected to be", name)
			} else if !tc.wantOk && ok {
				t.Fatalf("%s: queue is not empty when it was expected to be", name)
			}
			if ok && res != tc.want {
				t.Fatalf("%s: got %d, want %d", name, res, tc.want)
			}
			if len(tc.queue) != len(tc.wantLeft) {
				t.Fatalf("%s: tc.queue.len() = %d, want %d", name, len(tc.queue), len(tc.wantLeft))
			}
			for i, r := range tc.wantLeft {
				if r != tc.queue[i] {
					t.Errorf("%s: mismatch at index %d, %d != %d", name, i, r, tc.queue[i])
				}
			}
		})
	}
}

func TestPriorityQueueSortsElementsByPriority(t *testing.T) {
	tests := map[string]struct {
		queue priorityQueue[int]
		want  priorityQueue[int]
	}{
		"keeps an already sorted queue unchanged": {queue: priorityQueue[int]{1, 2, 3}, want: priorityQueue[int]{1, 2, 3}},
		"sorts a partially ordered queue":         {queue: priorityQueue[int]{2, 3, 1}, want: priorityQueue[int]{1, 2, 3}},
		"sorts a reverse-ordered queue":           {queue: priorityQueue[int]{5, 4, 3, 2, 1}, want: priorityQueue[int]{1, 2, 3, 4, 5}},
	}

	f := func(a int, b int) int {
		if a < b {
			return -1
		}
		return 1
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.queue.sortBy(f)

			for i, r := range tc.want {
				if r != tc.queue[i] {
					t.Errorf("%s: %d != %d", name, r, tc.queue[i])
				}
			}
		})
	}
}

func TestPriorityQueueInsertsElementsAtTheCorrectPosition(t *testing.T) {
	tests := map[string]struct {
		queue priorityQueue[int]
		input int
		want  priorityQueue[int]
	}{
		"inserts a duplicate at the front":  {queue: priorityQueue[int]{1, 2, 3}, input: 1, want: priorityQueue[int]{1, 1, 2, 3}},
		"inserts a new maximum at the end":  {queue: priorityQueue[int]{1, 2, 3}, input: 5, want: priorityQueue[int]{1, 2, 3, 5}},
		"inserts a duplicate in the middle": {queue: priorityQueue[int]{1, 2, 3}, input: 2, want: priorityQueue[int]{1, 2, 2, 3}},
	}

	f := func(a int, b int) bool {
		return a < b
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.queue.insertAt(tc.input, f)
			if len(tc.queue) != len(tc.want) {
				t.Fatalf("%s: mismatching array lengths,  %d != %d", name, len(tc.queue), len(tc.want))
			}

			for i, r := range tc.want {
				if r != tc.queue[i] {
					t.Errorf("%s: mismatch at index %d,  %d != %d", name, i, r, tc.queue[i])
				}
			}
		})
	}
}

func TestPriorityQueueFetchesTheFirstMatchingElement(t *testing.T) {
	tests := map[string]struct {
		queue    priorityQueue[int]
		matcher  func(int) bool
		want     int
		wantOk   bool
		wantLeft priorityQueue[int]
	}{
		"finds the first matching element": {
			queue:    priorityQueue[int]{1, 2, 3},
			matcher:  func(v int) bool { return v == 1 },
			want:     1,
			wantOk:   true,
			wantLeft: priorityQueue[int]{2, 3},
		},
		"finds a matching element in the middle": {
			queue:    priorityQueue[int]{1, 2, 3},
			matcher:  func(v int) bool { return v == 2 },
			want:     2,
			wantOk:   true,
			wantLeft: priorityQueue[int]{1, 3},
		},
		"returns not found when no element matches": {
			queue:    priorityQueue[int]{1, 2, 3},
			matcher:  func(v int) bool { return v == 4 },
			want:     0,
			wantOk:   false,
			wantLeft: priorityQueue[int]{1, 2, 3},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res, ok := tc.queue.fetchBy(tc.matcher)
			if ok != tc.wantOk {
				t.Fatalf("%s: got ok=%v, want %v", name, ok, tc.wantOk)
			}
			if ok && res != tc.want {
				t.Fatalf("%s: got %d, want %d", name, res, tc.want)
			}
			if len(tc.queue) != len(tc.wantLeft) {
				t.Fatalf("%s: mismatching array lengths, %d != %d", name, len(tc.queue), len(tc.wantLeft))
			}
			for i, r := range tc.wantLeft {
				if r != tc.queue[i] {
					t.Errorf("%s: mismatch at index %d, %d != %d", name, i, r, tc.queue[i])
				}
			}
		})
	}
}
