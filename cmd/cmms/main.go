package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robotfleetos/robotfleetos/internal/cmms"
)

func main() {
	cfg, err := cmms.LoadConfig("")
	if err != nil {
		log.Fatalf("cmms: load config: %v", err)
	}
	store := cmms.NewStore()
	fleet := cmms.NewFleetClient(cfg.Fleet.APIURL)
	svc := cmms.NewService(store, fleet)
	server := &cmms.Server{Service: svc}
	httpSrv := &http.Server{Addr: cfg.CMMS.Listen, Handler: server.Handler()}
	go func() {
		log.Printf("cmms: API on http://localhost%s", cfg.CMMS.Listen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("cmms: %v", err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("cmms: shutting down...")
	_ = httpSrv.Close()
	log.Println("cmms: done")
}
