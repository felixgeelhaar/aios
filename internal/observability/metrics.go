package observability

import "sync"

type Metrics struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewMetrics() *Metrics {
	return &Metrics{counts: map[string]int{}}
}

func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[name]++
}

func (m *Metrics) Value(name string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counts[name]
}
