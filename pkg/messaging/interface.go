// Package messaging defines the message bus abstraction used by all layers.
package messaging

import (
	"context"

	"github.com/robotfleetos/robotfleetos/pkg/api"
)

// Topic names (partitioned by area/zone in implementation).
const (
	TopicWorkOrders   = "fleet.work_orders"
	TopicZoneTasks    = "area.zone_tasks"
	TopicRobotCommands = "zone.robot_commands"
	TopicRobotStatus  = "edge.robot_status"
	TopicZoneSummary  = "zone.summary"
	TopicAreaSummary  = "area.summary"
)

// Publisher publishes messages to a topic (or partition).
type Publisher interface {
	Publish(ctx context.Context, topic string, key string, value []byte) error
}

// Subscriber subscribes to a topic (or partition) and receives messages.
type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler func(key string, value []byte) error) error
}

// Bus combines publish and subscribe; implementations can be Kafka, NATS JetStream, etc.
type Bus interface {
	Publisher
	Subscriber
}

// WorkOrderPublisher is used by the fleet layer.
type WorkOrderPublisher interface {
	PublishWorkOrder(ctx context.Context, order *api.WorkOrder) error
}

// ZoneTaskPublisher is used by the area layer.
type ZoneTaskPublisher interface {
	PublishZoneTask(ctx context.Context, task *api.ZoneTask) error
}

// RobotCommandPublisher is used by the zone layer.
type RobotCommandPublisher interface {
	PublishRobotCommand(ctx context.Context, cmd *api.RobotCommand) error
}
