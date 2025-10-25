package metric

import "sync"

// SimpleMovingAverage represents a simple moving average over a fixed-size window
type SimpleMovingAverage interface {
	// Observe adds a new value to the moving average
	Observe(v float64)
	// Load returns the current average
	Load() float64
}

// NewSimpleMovingAverage creates a new SimpleMovingAverage with the specified size
func NewSimpleMovingAverage(size int) SimpleMovingAverage {
	return &sma{
		window: make([]float64, size),
		size:   size,
	}
}

type sma struct {
	mu     sync.RWMutex
	window []float64
	size   int
	index  int
	sum    float64
	count  int // number of values added so far
}

func (s *sma) Observe(v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sum -= s.window[s.index]
	s.window[s.index] = v
	s.sum += v
	s.index = (s.index + 1) % s.size
	if s.count < s.size {
		s.count++
	}
}

func (s *sma) Load() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.count == 0 {
		return 0
	}
	return s.sum / float64(s.count)
}
