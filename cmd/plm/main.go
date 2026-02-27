package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robotfleetos/robotfleetos/internal/plm"
)

func main() {
	cfg, err := plm.LoadConfig("")
	if err != nil {
		log.Fatalf("plm: load config: %v", err)
	}
	store := plm.NewStore()
	svc := plm.NewService(store)
	server := &plm.Server{Service: svc}
	httpSrv := &http.Server{Addr: cfg.PLM.Listen, Handler: server.Handler()}
	go func() {
		log.Printf("plm: API on http://localhost%s", cfg.PLM.Listen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("plm: %v", err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("plm: shutting down...")
	_ = httpSrv.Close()
	log.Println("plm: done")
}
