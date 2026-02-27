// QMS (Quality Management System): inspections, NCRs, holds.
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robotfleetos/robotfleetos/internal/qms"
)

func main() {
	cfg, err := qms.LoadConfig("")
	if err != nil {
		log.Fatalf("qms: load config: %v", err)
	}

	store := qms.NewStore()
	svc := qms.NewService(store)
	server := &qms.Server{Service: svc}

	httpSrv := &http.Server{
		Addr:    cfg.QMS.Listen,
		Handler: server.Handler(),
	}

	go func() {
		log.Printf("qms: API on http://localhost%s", cfg.QMS.Listen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("qms: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("qms: shutting down...")
	_ = httpSrv.Close()
	log.Println("qms: done")
}
