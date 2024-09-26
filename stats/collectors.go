package stats

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const defaultPauseDur = 10 * time.Second

type gaugeTagsFunc = func(key string, tags Tags, val uint64)

type Collector interface {
	Collect(gaugeTagsFunc)
	Zero(gaugeTagsFunc)
	ID() string
}

type aggregatedCollector struct {
	c         map[string]Collector
	PauseDur  time.Duration
	gaugeFunc gaugeTagsFunc
	mu        sync.Mutex
}

func (p *aggregatedCollector) Add(c Collector) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.c == nil {
		p.c = make(map[string]Collector)
	}

	if _, ok := p.c[c.ID()]; ok {
		return fmt.Errorf("collector with ID %s already register", c.ID())
	}

	p.c[c.ID()] = c
	return nil
}

func (p *aggregatedCollector) Run(ctx context.Context) {
	defer p.allZero()
	p.allCollect()

	if p.PauseDur <= 0 {
		p.PauseDur = defaultPauseDur
	}

	tick := time.NewTicker(p.PauseDur)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			p.allCollect()
		}
	}
}

func (p *aggregatedCollector) allCollect() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.c {
		c.Collect(p.gaugeFunc)
	}
}

func (p *aggregatedCollector) allZero() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.c {
		c.Zero(p.gaugeFunc)
	}
}
