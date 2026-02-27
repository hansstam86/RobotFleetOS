package fleet

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/robotfleetos/robotfleetos/pkg/api"
	"github.com/robotfleetos/robotfleetos/pkg/messaging"
)

// GlobalState holds an eventually consistent view of the fleet (area summaries).
type GlobalState struct {
	mu     sync.RWMutex
	areas  map[api.AreaID]*api.AreaSummary
}

// NewGlobalState returns a new global state aggregator.
func NewGlobalState() *GlobalState {
	return &GlobalState{areas: make(map[api.AreaID]*api.AreaSummary)}
}

// Run subscribes to area summaries and updates internal state. Registers the handler and returns;
// the bus will invoke the handler whenever an area publishes a summary.
func (g *GlobalState) Run(ctx context.Context, bus messaging.Subscriber) error {
	return bus.Subscribe(ctx, messaging.TopicAreaSummary, func(key string, value []byte) error {
		var sum api.AreaSummary
		if err := json.Unmarshal(value, &sum); err != nil {
			return err
		}
		g.mu.Lock()
		g.areas[sum.AreaID] = &sum
		g.mu.Unlock()
		return nil
	})
}

// GetArea returns the latest summary for an area, or nil if unknown.
func (g *GlobalState) GetArea(id api.AreaID) *api.AreaSummary {
	g.mu.RLock()
	defer g.mu.RUnlock()
	s := g.areas[id]
	if s == nil {
		return nil
	}
	cp := *s
	return &cp
}

// GetAllAreas returns a copy of all area summaries.
func (g *GlobalState) GetAllAreas() []api.AreaSummary {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]api.AreaSummary, 0, len(g.areas))
	for _, s := range g.areas {
		out = append(out, *s)
	}
	return out
}

// TotalRobots returns the sum of robot counts across all areas (eventually consistent).
func (g *GlobalState) TotalRobots() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var n int
	for _, s := range g.areas {
		n += s.RobotCount
	}
	return n
}
