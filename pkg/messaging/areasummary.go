package messaging

import (
	"context"
	"encoding/json"

	"github.com/robotfleetos/robotfleetos/pkg/api"
)

// AreaSummaryPublisher publishes area summaries to the fleet (TopicAreaSummary).
type AreaSummaryPublisher struct {
	bus Publisher
}

// NewAreaSummaryPublisher returns a publisher for area summaries.
func NewAreaSummaryPublisher(bus Publisher) *AreaSummaryPublisher {
	return &AreaSummaryPublisher{bus: bus}
}

// PublishAreaSummary serializes the summary and publishes to TopicAreaSummary with key = area_id.
func (p *AreaSummaryPublisher) PublishAreaSummary(ctx context.Context, sum *api.AreaSummary) error {
	data, err := json.Marshal(sum)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, TopicAreaSummary, string(sum.AreaID), data)
}
