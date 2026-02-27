// Edge layer: one process per robot or small cell.
// Subscribes to robot commands, executes them (stub or real protocol), publishes robot status to the zone.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robotfleetos/robotfleetos/internal/edge"
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

	cfg, err := edge.LoadConfig("")
	if err != nil {
		log.Fatalf("edge: load config: %v", err)
	}

	busURL := os.Getenv("MESSAGING_URL")
	if busURL == "" {
		busURL = cfg.Messaging.Broker
	}
	bus, err := messaging.NewBusFromURL(busURL)
	if err != nil {
		log.Fatalf("edge: connect to message bus: %v", err)
	}
	statusPub := messaging.NewRobotStatusPublisher(bus)

	gw := edge.NewGateway(
		api.RobotID(cfg.Edge.RobotID),
		api.ZoneID(cfg.Edge.ZoneID),
		cfg.Edge.Protocol,
		statusPub,
		bus,
		2*time.Second,
		2*time.Second,
	)

	log.Printf("edge: starting %s (zone %s, protocol %s)", cfg.Edge.RobotID, cfg.Edge.ZoneID, cfg.Edge.Protocol)
	if err := gw.Run(ctx); err != nil && err != context.Canceled {
		log.Printf("edge: %v", err)
		os.Exit(1)
	}
	log.Println("edge: shutdown")
}
