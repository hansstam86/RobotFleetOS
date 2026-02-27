package area

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robotfleetos/robotfleetos/pkg/api"
	"github.com/robotfleetos/robotfleetos/pkg/messaging"
)

// Controller runs the area layer: consumes work orders, publishes zone tasks, aggregates zone summaries, reports to fleet.
type Controller struct {
	areaID   api.AreaID
	zones    []api.ZoneID
	workPub  messaging.WorkOrderPublisher
	zonePub  messaging.ZoneTaskPublisher
	areaPub  *messaging.AreaSummaryPublisher // publishes to fleet
	bus      messaging.Subscriber
	reportInterval time.Duration

	mu          sync.RWMutex
	zoneSummary map[api.ZoneID]*api.ZoneSummary
	taskSeq     atomic.Uint64
}

// NewController creates an area controller that consumes work orders from the bus and publishes zone tasks and area summaries.
func NewController(areaID api.AreaID, zones []api.ZoneID, zonePub messaging.ZoneTaskPublisher, areaPub *messaging.AreaSummaryPublisher, bus messaging.Subscriber, reportInterval time.Duration) *Controller {
	zoneMap := make(map[api.ZoneID]*api.ZoneSummary)
	for _, z := range zones {
		zoneMap[z] = nil
	}
	return &Controller{
		areaID:        areaID,
		zones:         zones,
		zonePub:       zonePub,
		areaPub:       areaPub,
		bus:           bus,
		reportInterval: reportInterval,
		zoneSummary:   zoneMap,
	}
}

// Run subscribes to work orders and zone summaries, and periodically publishes area summary. Blocks until ctx is done.
func (c *Controller) Run(ctx context.Context) error {
	// Subscribe to work orders: filter by our area_id, dispatch to zones.
	if err := c.bus.Subscribe(ctx, messaging.TopicWorkOrders, c.handleWorkOrder); err != nil {
		return err
	}
	// Subscribe to zone summaries: aggregate and update local state.
	if err := c.bus.Subscribe(ctx, messaging.TopicZoneSummary, c.handleZoneSummary); err != nil {
		return err
	}

	// Periodically publish area summary to fleet.
	ticker := time.NewTicker(c.reportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			c.publishAreaSummary(ctx)
		}
	}
}

func (c *Controller) handleWorkOrder(key string, value []byte) error {
	var order api.WorkOrder
	if err := json.Unmarshal(value, &order); err != nil {
		return err
	}
	if order.AreaID != c.areaID {
		return nil
	}
	// Dispatch to one zone (round-robin or first). For simplicity: first zone.
	if len(c.zones) == 0 {
		log.Printf("area %s: no zones configured, dropping work order %s", c.areaID, order.ID)
		return nil
	}
	zoneID := c.zones[int(c.taskSeq.Add(1))%len(c.zones)]
	task := &api.ZoneTask{
		ID:        api.TaskID(genTaskID(zoneID)),
		ZoneID:    zoneID,
		OrderID:   order.ID,
		Payload:   order.Payload,
		CreatedAt: time.Now().UTC(),
	}
	if err := c.zonePub.PublishZoneTask(context.Background(), task); err != nil {
		log.Printf("area %s: publish zone task: %v", c.areaID, err)
		return err
	}
	log.Printf("area %s: work order %s -> zone %s task %s", c.areaID, order.ID, zoneID, task.ID)
	return nil
}

func (c *Controller) handleZoneSummary(key string, value []byte) error {
	var sum api.ZoneSummary
	if err := json.Unmarshal(value, &sum); err != nil {
		return err
	}
	// Only care about zones we own.
	if !c.ownsZone(sum.ZoneID) {
		return nil
	}
	c.mu.Lock()
	c.zoneSummary[sum.ZoneID] = &sum
	c.mu.Unlock()
	return nil
}

func (c *Controller) ownsZone(z api.ZoneID) bool {
	for _, id := range c.zones {
		if id == z {
			return true
		}
	}
	return false
}

func (c *Controller) publishAreaSummary(ctx context.Context) {
	c.mu.RLock()
	zoneCount := 0
	robotCount := 0
	for _, s := range c.zoneSummary {
		if s != nil {
			zoneCount++
			robotCount += s.RobotCount
		}
	}
	c.mu.RUnlock()

	sum := &api.AreaSummary{
		AreaID:     c.areaID,
		ZoneCount:  zoneCount,
		RobotCount: robotCount,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := c.areaPub.PublishAreaSummary(ctx, sum); err != nil {
		log.Printf("area %s: publish area summary: %v", c.areaID, err)
	}
}

func genTaskID(zoneID api.ZoneID) string {
	return string(zoneID) + "-" + time.Now().Format("20060102150405")
}
