package messaging

import (
	"context"
	"encoding/json"

	"github.com/robotfleetos/robotfleetos/pkg/api"
)

// ZoneSummaryPublisher publishes zone summaries to the area (TopicZoneSummary).
type ZoneSummaryPublisher struct {
	bus Publisher
}

// NewZoneSummaryPublisher returns a publisher for zone summaries.
func NewZoneSummaryPublisher(bus Publisher) *ZoneSummaryPublisher {
	return &ZoneSummaryPublisher{bus: bus}
}

// PublishZoneSummary serializes the summary and publishes to TopicZoneSummary with key = zone_id.
func (p *ZoneSummaryPublisher) PublishZoneSummary(ctx context.Context, sum *api.ZoneSummary) error {
	data, err := json.Marshal(sum)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, TopicZoneSummary, string(sum.ZoneID), data)
}
