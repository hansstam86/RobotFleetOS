// Fleet layer: global scheduler, API, analytics.
// Scales via replicas; state in shared store + message bus.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robotfleetos/robotfleetos/internal/fleet"
	"github.com/robotfleetos/robotfleetos/pkg/messaging"
	"github.com/robotfleetos/robotfleetos/pkg/state"
)

func getBrokerURL(cfg *fleet.Config) string {
	if u := os.Getenv("MESSAGING_URL"); u != "" {
		return u
	}
	return cfg.Messaging.Broker
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	cfg, err := fleet.LoadConfig("")
	if err != nil {
		log.Fatalf("fleet: load config: %v", err)
	}

	busURL := getBrokerURL(cfg)
	bus, busErr := messaging.NewBusFromURL(busURL)
	if busErr != nil {
		log.Fatalf("fleet: connect to message bus: %v", busErr)
	}
	_ = state.NewMemoryStore() // reserved for fleet-level state (e.g. work order ledger)

	workOrderPub := messaging.NewWorkOrderPublisher(bus)
	scheduler := fleet.NewScheduler(workOrderPub)
	globalState := fleet.NewGlobalState()

	// Register handler for area summaries (updates global state when areas report in).
	go func() {
		_ = globalState.Run(ctx, bus)
	}()

	server := &fleet.Server{Scheduler: scheduler, State: globalState}

	srv := &http.Server{Addr: cfg.Fleet.APIListen, Handler: server.Handler()}
	go func() {
		log.Printf("fleet: API listening on %s", cfg.Fleet.APIListen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("fleet: HTTP server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("fleet: shutting down")
	_ = srv.Shutdown(context.Background())
	log.Println("fleet: shutdown complete")
}
