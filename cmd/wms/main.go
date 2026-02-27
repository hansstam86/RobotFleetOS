// WMS (Warehouse Management System): locations, inventory, pick/putaway/move tasks, release to Fleet.
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robotfleetos/robotfleetos/internal/wms"
)

func main() {
	cfg, err := wms.LoadConfig("")
	if err != nil {
		log.Fatalf("wms: load config: %v", err)
	}

	store := wms.NewStore()
	fleetClient := wms.NewFleetClient(cfg.Fleet.APIURL)
	svc := wms.NewService(store, fleetClient, cfg.WMS.WarehouseAreaID)
	server := &wms.Server{Service: svc}

	httpSrv := &http.Server{
		Addr:    cfg.WMS.Listen,
		Handler: server.Handler(),
	}

	go func() {
		log.Printf("wms: API on http://localhost%s (Fleet: %s)", cfg.WMS.Listen, cfg.Fleet.APIURL)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("wms: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("wms: shutting down...")
	_ = httpSrv.Close()
	log.Println("wms: done")
}
