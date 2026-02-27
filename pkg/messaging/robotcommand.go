package messaging

import (
	"context"
	"encoding/json"

	"github.com/robotfleetos/robotfleetos/pkg/api"
)

// robotCommandPublisher publishes robot commands to the bus (TopicRobotCommands, key = RobotID).
type robotCommandPublisher struct {
	bus Publisher
}

// NewRobotCommandPublisher returns a RobotCommandPublisher that serializes commands to JSON and publishes to TopicRobotCommands.
func NewRobotCommandPublisher(bus Publisher) *robotCommandPublisher {
	return &robotCommandPublisher{bus: bus}
}

// PublishRobotCommand serializes the command and publishes with key = cmd.RobotID for partitioning.
func (p *robotCommandPublisher) PublishRobotCommand(ctx context.Context, cmd *api.RobotCommand) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, TopicRobotCommands, string(cmd.RobotID), data)
}
