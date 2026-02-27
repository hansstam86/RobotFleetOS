package edge

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/robotfleetos/robotfleetos/pkg/api"
	"github.com/robotfleetos/robotfleetos/pkg/messaging"
)

// Simulator simulates many robots in one process: each defers firmware when BUSY and applies when IDLE.
type Simulator struct {
	zoneID        api.ZoneID
	robots        []api.RobotID
	statusPub     *messaging.RobotStatusPublisher
	bus           messaging.Subscriber
	statusInterval time.Duration
	taskDuration   time.Duration

	mu     sync.RWMutex
	state  map[api.RobotID]*robotSimState
}

type robotSimState struct {
	state               string
	modelID             string
	firmwareVersion     string
	firmwareUpdateStatus string
	pendingFirmware     *api.RobotCommand
}

// NewSimulator creates a simulator for the given robot IDs in the zone.
func NewSimulator(
	zoneID api.ZoneID,
	robots []api.RobotID,
	statusPub *messaging.RobotStatusPublisher,
	bus messaging.Subscriber,
	statusInterval time.Duration,
	taskDuration time.Duration,
) *Simulator {
	state := make(map[api.RobotID]*robotSimState, len(robots))
	for _, r := range robots {
		state[r] = &robotSimState{
			state:               "IDLE",
			modelID:             "stub-model",
			firmwareVersion:     "1.0.0",
			firmwareUpdateStatus: api.FirmwareStatusIdle,
		}
	}
	return &Simulator{
		zoneID:         zoneID,
		robots:         robots,
		statusPub:      statusPub,
		bus:            bus,
		statusInterval: statusInterval,
		taskDuration:   taskDuration,
		state:          state,
	}
}

// Run subscribes to commands and publishes status for all robots. Blocks until ctx is done.
func (s *Simulator) Run(ctx context.Context) error {
	if err := s.bus.Subscribe(ctx, messaging.TopicRobotCommands, s.handleCommand); err != nil {
		return err
	}
	ticker := time.NewTicker(s.statusInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.publishAllStatus(ctx)
		}
	}
}

func (s *Simulator) handleCommand(key string, value []byte) error {
	var cmd api.RobotCommand
	if err := json.Unmarshal(value, &cmd); err != nil {
		return err
	}
	s.mu.RLock()
	_, ok := s.state[cmd.RobotID]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	switch cmd.Type {
	case api.RobotCommandTypeFirmwareUpdate:
		s.handleFirmwareUpdate(cmd)
	default:
		s.handleTask(cmd)
	}
	return nil
}

func (s *Simulator) handleTask(cmd api.RobotCommand) {
	s.mu.Lock()
	st := s.state[cmd.RobotID]
	if st == nil {
		s.mu.Unlock()
		return
	}
	st.state = "BUSY"
	s.mu.Unlock()
	go func() {
		time.Sleep(s.taskDuration)
		s.mu.Lock()
		st := s.state[cmd.RobotID]
		if st != nil && st.state == "BUSY" {
			st.state = "IDLE"
			pending := st.pendingFirmware
			st.pendingFirmware = nil
			s.mu.Unlock()
			if pending != nil {
				s.handleFirmwareUpdate(*pending)
			}
		} else {
			s.mu.Unlock()
		}
	}()
}

func (s *Simulator) handleFirmwareUpdate(cmd api.RobotCommand) {
	var payload api.FirmwareUpdatePayload
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		return
	}
	s.mu.Lock()
	st := s.state[cmd.RobotID]
	if st == nil {
		s.mu.Unlock()
		return
	}
	if payload.ModelID != "" && payload.ModelID != st.modelID {
		s.mu.Unlock()
		return
	}
	if st.state == "BUSY" {
		st.pendingFirmware = &cmd
		s.mu.Unlock()
		return
	}
	st.state = "BUSY"
	st.firmwareUpdateStatus = api.FirmwareStatusDownloading
	s.mu.Unlock()
	go func() {
		time.Sleep(2 * time.Second)
		s.mu.Lock()
		st := s.state[cmd.RobotID]
		if st != nil {
			st.firmwareUpdateStatus = api.FirmwareStatusApplying
		}
		s.mu.Unlock()
		time.Sleep(2 * time.Second)
		s.mu.Lock()
		st = s.state[cmd.RobotID]
		if st != nil {
			st.state = "IDLE"
			st.firmwareVersion = payload.Version
			st.firmwareUpdateStatus = api.FirmwareStatusSuccess
		}
		s.mu.Unlock()
	}()
}

func (s *Simulator) publishAllStatus(ctx context.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, robotID := range s.robots {
		st := s.state[robotID]
		if st == nil {
			continue
		}
		status := &api.RobotStatus{
			RobotID:   robotID,
			State:     st.state,
			Battery:   100,
			UpdatedAt: time.Now().UTC(),
			Extra: map[string]interface{}{
				api.ExtraModelID:             st.modelID,
				api.ExtraFirmwareVersion:     st.firmwareVersion,
				api.ExtraFirmwareUpdateStatus: st.firmwareUpdateStatus,
			},
		}
		_ = s.statusPub.PublishRobotStatus(ctx, status)
	}
}
