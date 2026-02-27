package edge

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/robotfleetos/robotfleetos/pkg/api"
	"github.com/robotfleetos/robotfleetos/pkg/messaging"
)

// Gateway is the edge gateway for one robot: consumes commands, executes them (stub or real protocol), publishes status.
type Gateway struct {
	robotID   api.RobotID
	zoneID    api.ZoneID
	protocol  string
	statusPub *messaging.RobotStatusPublisher
	bus       messaging.Subscriber
	statusInterval time.Duration
	taskDuration   time.Duration // how long to stay BUSY for a TASK (stub)

	mu    sync.RWMutex
	state string // IDLE, BUSY, ERROR, CHARGING
	battery float64
	// Firmware (simulation)
	modelID             string
	firmwareVersion     string
	firmwareUpdateStatus string
	pendingFirmware     *api.RobotCommand // applied when robot becomes IDLE
}

// NewGateway creates an edge gateway for the given robot.
func NewGateway(
	robotID api.RobotID,
	zoneID api.ZoneID,
	protocol string,
	statusPub *messaging.RobotStatusPublisher,
	bus messaging.Subscriber,
	statusInterval time.Duration,
	taskDuration time.Duration,
) *Gateway {
	if statusInterval <= 0 {
		statusInterval = 2 * time.Second
	}
	if taskDuration <= 0 {
		taskDuration = 2 * time.Second
	}
	return &Gateway{
		robotID:             robotID,
		zoneID:              zoneID,
		protocol:            protocol,
		statusPub:           statusPub,
		bus:                 bus,
		statusInterval:      statusInterval,
		taskDuration:        taskDuration,
		state:               "IDLE",
		battery:             100,
		modelID:             "stub-model",
		firmwareVersion:     "1.0.0",
		firmwareUpdateStatus: api.FirmwareStatusIdle,
	}
}

// Run subscribes to robot commands and periodically publishes status. Blocks until ctx is done.
func (g *Gateway) Run(ctx context.Context) error {
	if err := g.bus.Subscribe(ctx, messaging.TopicRobotCommands, g.handleCommand); err != nil {
		return err
	}

	ticker := time.NewTicker(g.statusInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			g.publishStatus(ctx)
		}
	}
}

func (g *Gateway) handleCommand(key string, value []byte) error {
	var cmd api.RobotCommand
	if err := json.Unmarshal(value, &cmd); err != nil {
		return err
	}
	if cmd.RobotID != g.robotID {
		return nil
	}
	log.Printf("edge %s: received command %s type=%s", g.robotID, cmd.ID, cmd.Type)
	switch cmd.Type {
	case api.RobotCommandTypeFirmwareUpdate:
		g.simulateFirmwareUpdate(cmd)
	default:
		switch g.protocol {
		case "stub":
			g.executeStub(cmd)
		default:
			g.executeStub(cmd)
		}
	}
	return nil
}

func (g *Gateway) executeStub(cmd api.RobotCommand) {
	g.mu.Lock()
	g.state = "BUSY"
	g.mu.Unlock()
	go func() {
		time.Sleep(g.taskDuration)
		g.mu.Lock()
		if g.state == "BUSY" {
			g.state = "IDLE"
		}
		var pending *api.RobotCommand
		if g.pendingFirmware != nil {
			pending = g.pendingFirmware
			g.pendingFirmware = nil
		}
		g.mu.Unlock()
		log.Printf("edge %s: task %s completed (stub)", g.robotID, cmd.ID)
		if pending != nil {
			log.Printf("edge %s: applying deferred firmware update", g.robotID)
			g.simulateFirmwareUpdate(*pending)
		}
	}()
}

func (g *Gateway) simulateFirmwareUpdate(cmd api.RobotCommand) {
	var payload api.FirmwareUpdatePayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		log.Printf("edge %s: invalid firmware payload: %v", g.robotID, err)
		return
	}
	if payload.ModelID != "" && payload.ModelID != g.modelID {
		log.Printf("edge %s: skip firmware (model %s != %s)", g.robotID, payload.ModelID, g.modelID)
		return
	}
	g.mu.Lock()
	if g.state == "BUSY" {
		g.pendingFirmware = &cmd
		g.mu.Unlock()
		log.Printf("edge %s: firmware update deferred until work order complete", g.robotID)
		return
	}
	g.mu.Unlock()
	go func() {
		g.mu.Lock()
		g.state = "BUSY"
		g.firmwareUpdateStatus = api.FirmwareStatusDownloading
		g.mu.Unlock()
		log.Printf("edge %s: firmware simulating download -> %s", g.robotID, payload.Version)
		time.Sleep(2 * time.Second)
		g.mu.Lock()
		g.firmwareUpdateStatus = api.FirmwareStatusApplying
		g.mu.Unlock()
		time.Sleep(2 * time.Second)
		g.mu.Lock()
		g.state = "IDLE"
		g.firmwareVersion = payload.Version
		g.firmwareUpdateStatus = api.FirmwareStatusSuccess
		g.mu.Unlock()
		log.Printf("edge %s: firmware update complete -> %s", g.robotID, payload.Version)
	}()
}

func (g *Gateway) publishStatus(ctx context.Context) {
	g.mu.RLock()
	state := g.state
	battery := g.battery
	modelID := g.modelID
	fwVer := g.firmwareVersion
	fwStatus := g.firmwareUpdateStatus
	g.mu.RUnlock()

	status := &api.RobotStatus{
		RobotID:   g.robotID,
		State:     state,
		Battery:   battery,
		UpdatedAt: time.Now().UTC(),
		Extra: map[string]interface{}{
			api.ExtraModelID:             modelID,
			api.ExtraFirmwareVersion:     fwVer,
			api.ExtraFirmwareUpdateStatus: fwStatus,
		},
	}
	if err := g.statusPub.PublishRobotStatus(ctx, status); err != nil {
		log.Printf("edge %s: publish status: %v", g.robotID, err)
	}
}
