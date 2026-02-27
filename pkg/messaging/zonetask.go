package messaging

import (
	"context"
	"encoding/json"

	"github.com/robotfleetos/robotfleetos/pkg/api"
)

// zoneTaskPublisher publishes zone tasks to the bus (TopicZoneTasks, key = ZoneID).
type zoneTaskPublisher struct {
	bus Publisher
}

// NewZoneTaskPublisher returns a ZoneTaskPublisher that serializes tasks to JSON and publishes to TopicZoneTasks.
func NewZoneTaskPublisher(bus Publisher) *zoneTaskPublisher {
	return &zoneTaskPublisher{bus: bus}
}

// PublishZoneTask serializes the zone task and publishes it with key = task.ZoneID for partitioning.
func (p *zoneTaskPublisher) PublishZoneTask(ctx context.Context, task *api.ZoneTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, TopicZoneTasks, string(task.ZoneID), data)
}
