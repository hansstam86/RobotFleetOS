// Area layer: one process per factory area (10k–100k robots).
// Subscribes to work orders from fleet, dispatches to zones, aggregates zone summaries, reports to fleet.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robotfleetos/robotfleetos/internal/area"
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

	cfg, err := area.LoadConfig("")
	if err != nil {
		log.Fatalf("area: load config: %v", err)
	}

	busURL := os.Getenv("MESSAGING_URL")
	if busURL == "" {
		busURL = cfg.Messaging.Broker
	}
	bus, err := messaging.NewBusFromURL(busURL)
	if err != nil {
		log.Fatalf("area: connect to message bus: %v", err)
	}
	zonePub := messaging.NewZoneTaskPublisher(bus)
	areaPub := messaging.NewAreaSummaryPublisher(bus)

	zones := make([]api.ZoneID, 0, len(cfg.Area.Zones))
	for _, z := range cfg.Area.Zones {
		zones = append(zones, api.ZoneID(z))
	}

	ctrl := area.NewController(
		api.AreaID(cfg.Area.AreaID),
		zones,
		zonePub,
		areaPub,
		bus,
		10*time.Second,
	)

	log.Printf("area: starting %s (zones: %v)", cfg.Area.AreaID, cfg.Area.Zones)
	if err := ctrl.Run(ctx); err != nil && err != context.Canceled {
		log.Printf("area: %v", err)
		os.Exit(1)
	}
	log.Println("area: shutdown")
}
