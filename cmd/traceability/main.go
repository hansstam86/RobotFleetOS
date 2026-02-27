// Traceability: serial/lot genealogy and recall reporting.
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robotfleetos/robotfleetos/internal/traceability"
)

func main() {
	cfg, err := traceability.LoadConfig("")
	if err != nil {
		log.Fatalf("traceability: load config: %v", err)
	}

	store := traceability.NewStore()
	svc := traceability.NewService(store)
	server := &traceability.Server{Service: svc}

	httpSrv := &http.Server{
		Addr:    cfg.Traceability.Listen,
		Handler: server.Handler(),
	}

	go func() {
		log.Printf("traceability: API on http://localhost%s", cfg.Traceability.Listen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("traceability: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("traceability: shutting down...")
	_ = httpSrv.Close()
	log.Println("traceability: done")
}
