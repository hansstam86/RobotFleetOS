package messaging

import (
	"context"
	"encoding/json"

	"github.com/robotfleetos/robotfleetos/pkg/api"
)

// RobotStatusPublisher publishes robot status to the zone (TopicRobotStatus).
type RobotStatusPublisher struct {
	bus Publisher
}

// NewRobotStatusPublisher returns a publisher for robot status.
func NewRobotStatusPublisher(bus Publisher) *RobotStatusPublisher {
	return &RobotStatusPublisher{bus: bus}
}

// PublishRobotStatus serializes the status and publishes to TopicRobotStatus with key = robot_id.
func (p *RobotStatusPublisher) PublishRobotStatus(ctx context.Context, status *api.RobotStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, TopicRobotStatus, string(status.RobotID), data)
}
