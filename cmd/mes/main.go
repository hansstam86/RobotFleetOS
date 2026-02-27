// MES (Manufacturing Execution System): production orders, release to Fleet, status, scrap.
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robotfleetos/robotfleetos/internal/mes"
)

func main() {
	cfg, err := mes.LoadConfig("")
	if err != nil {
		log.Fatalf("mes: load config: %v", err)
	}

	store := mes.NewStore()
	fleetClient := mes.NewFleetClient(cfg.Fleet.APIURL)
	svc := mes.NewService(store, fleetClient)
	server := &mes.Server{Service: svc}

	httpSrv := &http.Server{
		Addr:    cfg.MES.Listen,
		Handler: server.Handler(),
	}

	go func() {
		log.Printf("mes: API on http://localhost%s (Fleet: %s)", cfg.MES.Listen, cfg.Fleet.APIURL)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("mes: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("mes: shutting down...")
	_ = httpSrv.Close()
	log.Println("mes: done")
}
