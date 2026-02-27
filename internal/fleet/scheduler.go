package fleet

import (
	"context"
	"sync/atomic"

	"github.com/robotfleetos/robotfleetos/pkg/api"
	"github.com/robotfleetos/robotfleetos/pkg/messaging"
)

// Scheduler assigns work orders to areas by publishing them to the message bus.
// Area controllers subscribe to work orders (partitioned by AreaID) and pull their assignments.
type Scheduler struct {
	publisher messaging.WorkOrderPublisher
	seq       atomic.Uint64
}

// NewScheduler returns a scheduler that publishes to the given WorkOrderPublisher.
func NewScheduler(pub messaging.WorkOrderPublisher) *Scheduler {
	return &Scheduler{publisher: pub}
}

// SubmitWorkOrder publishes the work order to the bus. The order must have a non-empty AreaID.
// If ID is empty, a generated ID is assigned before publishing.
func (s *Scheduler) SubmitWorkOrder(ctx context.Context, order *api.WorkOrder) error {
	if order.ID == "" {
		order.ID = api.WorkOrderID(generateID("wo", s.seq.Add(1)))
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now()
	}
	return s.publisher.PublishWorkOrder(ctx, order)
}

// CancelWorkOrder can be implemented later by publishing a cancel message to a dedicated topic
// or by area/zone handling cancel by work order ID. For now we only support submit.
func (s *Scheduler) CancelWorkOrder(ctx context.Context, id api.WorkOrderID) error {
	// TODO: publish cancel event; areas/zones will stop work for this order
	_ = id
	return nil
}
