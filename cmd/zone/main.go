// Zone layer: one process per physical zone (100–2k robots).
// Receives zone tasks from area, dispatches robot commands, aggregates robot status, reports zone summary to area.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robotfleetos/robotfleetos/internal/zone"
	"github.com/robotfleetos/robotfleetos/pkg/api"
	"github.com/robotfleetos/robotfleetos/pkg/messaging"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	cfg, err := zone.LoadConfig("")
	if err != nil {
		log.Fatalf("zone: load config: %v", err)
	}

	busURL := os.Getenv("MESSAGING_URL")
	if busURL == "" {
		busURL = cfg.Messaging.Broker
	}
	bus, err := messaging.NewBusFromURL(busURL)
	if err != nil {
		log.Fatalf("zone: connect to message bus: %v", err)
	}
	cmdPub := messaging.NewRobotCommandPublisher(bus)
	summaryPub := messaging.NewZoneSummaryPublisher(bus)

	robots := make([]api.RobotID, 0, len(cfg.Zone.Robots))
	for _, r := range cfg.Zone.Robots {
		robots = append(robots, api.RobotID(r))
	}

	ctrl := zone.NewController(
		api.ZoneID(cfg.Zone.ZoneID),
		robots,
		cmdPub,
		summaryPub,
		bus,
		5*time.Second,
	)

	log.Printf("zone: starting %s (robots: %v)", cfg.Zone.ZoneID, cfg.Zone.Robots)
	if err := ctrl.Run(ctx); err != nil && err != context.Canceled {
		log.Printf("zone: %v", err)
		os.Exit(1)
	}
	log.Println("zone: shutdown")
}
