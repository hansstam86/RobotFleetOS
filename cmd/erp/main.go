package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robotfleetos/robotfleetos/internal/erp"
)

func main() {
	cfg, err := erp.LoadConfig("")
	if err != nil {
		log.Fatalf("erp: load config: %v", err)
	}
	store := erp.NewStore()
	mes := erp.NewMESClient(cfg.MES.APIURL)
	svc := erp.NewService(store, mes, cfg.MES.DefaultAreaID)
	server := &erp.Server{Service: svc}
	httpSrv := &http.Server{Addr: cfg.ERP.Listen, Handler: server.Handler()}
	go func() {
		log.Printf("erp: API on http://localhost%s", cfg.ERP.Listen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erp: %v", err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("erp: shutting down...")
	_ = httpSrv.Close()
	log.Println("erp: done")
}
