package stats

import (
	"context"
	"sync"
	"time"
)

type gaugeTagsFunc = func(key string, tags Tags, val uint64)

type Collector interface {
	Collect(gaugeTagsFunc)
	Zero(gaugeTagsFunc)
}

type aggregatedCollector struct {
	c         []Collector
	PauseDur  time.Duration
	gaugeFunc gaugeTagsFunc
	mu        sync.Mutex
}

func (p *aggregatedCollector) Add(c Collector) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.c = append(p.c, c)
}

func (p *aggregatedCollector) Run(ctx context.Context) {
	defer p.allZero()
	p.allCollect()

	if p.PauseDur <= 0 {
		p.PauseDur = 10 * time.Second
		return
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
