package zone

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

// Controller runs the zone layer: consumes zone tasks, publishes robot commands, aggregates robot status, reports zone summary to area.
type Controller struct {
	zoneID   api.ZoneID
	robots   []api.RobotID
	cmdPub   messaging.RobotCommandPublisher
	summaryPub *messaging.ZoneSummaryPublisher
	bus      messaging.Subscriber
	reportInterval time.Duration

	mu          sync.RWMutex
	robotStatus map[api.RobotID]*api.RobotStatus
	cmdSeq      atomic.Uint64
}

// NewController creates a zone controller.
func NewController(
	zoneID api.ZoneID,
	robots []api.RobotID,
	cmdPub messaging.RobotCommandPublisher,
	summaryPub *messaging.ZoneSummaryPublisher,
	bus messaging.Subscriber,
	reportInterval time.Duration,
) *Controller {
	statusMap := make(map[api.RobotID]*api.RobotStatus)
	for _, r := range robots {
		statusMap[r] = nil
	}
	return &Controller{
		zoneID:        zoneID,
		robots:        robots,
		cmdPub:        cmdPub,
		summaryPub:    summaryPub,
		bus:           bus,
		reportInterval: reportInterval,
		robotStatus:   statusMap,
	}
}

// Run subscribes to zone tasks and robot status, and periodically publishes zone summary. Blocks until ctx is done.
func (c *Controller) Run(ctx context.Context) error {
	if err := c.bus.Subscribe(ctx, messaging.TopicZoneTasks, c.handleZoneTask); err != nil {
		return err
	}
	if err := c.bus.Subscribe(ctx, messaging.TopicRobotStatus, c.handleRobotStatus); err != nil {
		return err
	}

	ticker := time.NewTicker(c.reportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			c.publishZoneSummary(ctx)
		}
	}
}

func (c *Controller) handleZoneTask(key string, value []byte) error {
	var task api.ZoneTask
	if err := json.Unmarshal(value, &task); err != nil {
		return err
	}
	if task.ZoneID != c.zoneID {
		return nil
	}
	if len(c.robots) == 0 {
		log.Printf("zone %s: no robots, dropping task %s", c.zoneID, task.ID)
		return nil
	}
	cmdType := "TASK"
	if len(task.Payload) > 0 {
		var maybeFw struct {
			Type       string `json:"type"`
			CampaignID string `json:"campaign_id"`
			Version    string `json:"version"`
		}
		if json.Unmarshal(task.Payload, &maybeFw) == nil && (maybeFw.Type == "firmware_update" || (maybeFw.CampaignID != "" && maybeFw.Version != "")) {
			cmdType = api.RobotCommandTypeFirmwareUpdate
		}
	}

	if cmdType == api.RobotCommandTypeFirmwareUpdate {
		// Broadcast firmware to all robots in zone; each may apply when IDLE (busy robots defer).
		for _, robotID := range c.robots {
			cmd := &api.RobotCommand{
				ID:        api.TaskID(string(task.ID) + "-" + string(robotID)),
				RobotID:   robotID,
				Type:      api.RobotCommandTypeFirmwareUpdate,
				Payload:   task.Payload,
				CreatedAt: time.Now().UTC(),
			}
			if err := c.cmdPub.PublishRobotCommand(context.Background(), cmd); err != nil {
				log.Printf("zone %s: publish robot command %s: %v", c.zoneID, robotID, err)
				return err
			}
		}
		log.Printf("zone %s: firmware task %s -> %d robots (broadcast)", c.zoneID, task.ID, len(c.robots))
		return nil
	}

	robotID := c.robots[int(c.cmdSeq.Add(1))%len(c.robots)]
	cmd := &api.RobotCommand{
		ID:        task.ID,
		RobotID:   robotID,
		Type:      cmdType,
		Payload:   task.Payload,
		CreatedAt: time.Now().UTC(),
	}
	if err := c.cmdPub.PublishRobotCommand(context.Background(), cmd); err != nil {
		log.Printf("zone %s: publish robot command: %v", c.zoneID, err)
		return err
	}
	log.Printf("zone %s: task %s -> robot %s", c.zoneID, task.ID, robotID)
	return nil
}

func (c *Controller) handleRobotStatus(key string, value []byte) error {
	var status api.RobotStatus
	if err := json.Unmarshal(value, &status); err != nil {
		return err
	}
	if !c.ownsRobot(status.RobotID) {
		return nil
	}
	c.mu.Lock()
	c.robotStatus[status.RobotID] = &status
	c.mu.Unlock()
	return nil
}

func (c *Controller) ownsRobot(r api.RobotID) bool {
	for _, id := range c.robots {
		if id == r {
			return true
		}
	}
	return false
}

func (c *Controller) publishZoneSummary(ctx context.Context) {
	c.mu.RLock()
	robotCount := len(c.robots)
	healthy := 0
	busy := 0
	for _, s := range c.robotStatus {
		if s != nil {
			if s.State != "ERROR" {
				healthy++
			}
			if s.State == "BUSY" {
				busy++
			}
		}
	}
	c.mu.RUnlock()

	sum := &api.ZoneSummary{
		ZoneID:     c.zoneID,
		RobotCount: robotCount,
		Healthy:    healthy,
		Busy:       busy,
		UpdatedAt:  time.Now().UTC(),
	}
	if err := c.summaryPub.PublishZoneSummary(ctx, sum); err != nil {
		log.Printf("zone %s: publish zone summary: %v", c.zoneID, err)
	}
}
