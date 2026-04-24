package metrics

import "sync"

type boundedQueue struct {
	mu      sync.Mutex
	items   []Metric
	maxSize int
	dropped uint64
}

func newBoundedQueue(maxSize int) *boundedQueue {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &boundedQueue{
		items:   make([]Metric, 0, maxSize),
		maxSize: maxSize,
	}
}

func (q *boundedQueue) enqueue(m Metric) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.maxSize {
		// Drop oldest to keep latest behavior during pressure.
		q.items = q.items[1:]
		q.dropped++
	}
	q.items = append(q.items, m)
}

func (q *boundedQueue) popN(n int) []Metric {
	q.mu.Lock()
	defer q.mu.Unlock()

	if n <= 0 || len(q.items) == 0 {
		return nil
	}
	if n > len(q.items) {
		n = len(q.items)
	}
	out := append([]Metric(nil), q.items[:n]...)
	q.items = q.items[n:]
	return out
}

func (q *boundedQueue) prepend(items []Metric) {
	if len(items) == 0 {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(append([]Metric(nil), items...), q.items...)
	if len(q.items) > q.maxSize {
		excess := len(q.items) - q.maxSize
		q.items = q.items[excess:]
		q.dropped += uint64(excess)
	}
}

func (q *boundedQueue) len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *boundedQueue) droppedCount() uint64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.dropped
}
