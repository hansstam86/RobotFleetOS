package messaging

import (
	"context"
	"encoding/json"

	"github.com/robotfleetos/robotfleetos/pkg/api"
)

// WorkOrderPublisher publishes work orders to the bus (TopicWorkOrders, key = AreaID).
type workOrderPublisher struct {
	bus Publisher
}

// NewWorkOrderPublisher returns a WorkOrderPublisher that serializes orders to JSON and publishes to TopicWorkOrders.
func NewWorkOrderPublisher(bus Publisher) *workOrderPublisher {
	return &workOrderPublisher{bus: bus}
}

// PublishWorkOrder serializes the work order and publishes it with key = order.AreaID for partitioning.
func (p *workOrderPublisher) PublishWorkOrder(ctx context.Context, order *api.WorkOrder) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return p.bus.Publish(ctx, TopicWorkOrders, string(order.AreaID), data)
}
