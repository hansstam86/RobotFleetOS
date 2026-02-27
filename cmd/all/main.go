// All-in-one dev server: runs fleet + area + zone + edge (one per robot) in one process with a shared in-memory bus.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/robotfleetos/robotfleetos/internal/area"
	"github.com/robotfleetos/robotfleetos/internal/edge"
	"github.com/robotfleetos/robotfleetos/internal/fleet"
	"github.com/robotfleetos/robotfleetos/internal/zone"
	"github.com/robotfleetos/robotfleetos/pkg/api"
	"github.com/robotfleetos/robotfleetos/pkg/messaging"
	"github.com/robotfleetos/robotfleetos/pkg/state"
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

	// Single shared bus for fleet and area (in-memory for dev).
	bus := messaging.NewMemoryBus()
	_ = state.NewMemoryStore()

	// ---- Fleet ----
	fleetCfg, err := fleet.LoadConfig("")
	if err != nil || fleetCfg == nil {
		fleetCfg = &fleet.Config{}
		fleetCfg.Fleet.APIListen = ":8080"
	}
	workOrderPub := messaging.NewWorkOrderPublisher(bus)
	scheduler := fleet.NewScheduler(workOrderPub)
	globalState := fleet.NewGlobalState()
	go func() { _ = globalState.Run(ctx, bus) }()
	fleetServer := &fleet.Server{Scheduler: scheduler, State: globalState}
	httpSrv := &http.Server{Addr: fleetCfg.Fleet.APIListen, Handler: fleetServer.Handler()}
	go func() {
		log.Printf("fleet: API on http://localhost%s", fleetCfg.Fleet.APIListen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("fleet: %v", err)
		}
	}()

	// ---- Area (same bus) ----
	areaCfg, errArea := area.LoadConfig("")
	if errArea != nil || areaCfg == nil {
		areaCfg = &area.Config{}
		areaCfg.Area.AreaID = "area-1"
		areaCfg.Area.Zones = []string{"zone-1"}
	}
	zonePub := messaging.NewZoneTaskPublisher(bus)
	areaPub := messaging.NewAreaSummaryPublisher(bus)
	zones := make([]api.ZoneID, len(areaCfg.Area.Zones))
	for i, z := range areaCfg.Area.Zones {
		zones[i] = api.ZoneID(z)
	}
	areaCtrl := area.NewController(
		api.AreaID(areaCfg.Area.AreaID),
		zones,
		zonePub,
		areaPub,
		bus,
		10*time.Second,
	)
	go func() {
		log.Printf("area: %s (zones: %v)", areaCfg.Area.AreaID, areaCfg.Area.Zones)
		_ = areaCtrl.Run(ctx)
	}()

	// ---- Zone (same bus) ----
	zoneCfg, errZone := zone.LoadConfig("")
	if errZone != nil || zoneCfg == nil {
		zoneCfg = &zone.Config{}
		zoneCfg.Zone.ZoneID = "zone-1"
		zoneCfg.Zone.AreaID = "area-1"
		zoneCfg.Zone.Robots = []string{"robot-1"}
	}
	zoneRobots := make([]api.RobotID, len(zoneCfg.Zone.Robots))
	for i, r := range zoneCfg.Zone.Robots {
		zoneRobots[i] = api.RobotID(r)
	}
	// Override with SIMULATE_ROBOTS for large-scale firmware sim (e.g. 1200 robots).
	if nStr := os.Getenv("SIMULATE_ROBOTS"); nStr != "" {
		if n, err := strconv.Atoi(nStr); err == nil && n > 0 && n <= 1000000 {
			zoneRobots = make([]api.RobotID, n)
			zoneCfg.Zone.Robots = make([]string, n)
			for i := 0; i < n; i++ {
				id := api.RobotID(fmt.Sprintf("robot-%d", i+1))
				zoneRobots[i] = id
				zoneCfg.Zone.Robots[i] = string(id)
			}
			log.Printf("simulation: using %d simulated robots (SIMULATE_ROBOTS=%s)", n, nStr)
		}
	}
	cmdPub := messaging.NewRobotCommandPublisher(bus)
	zoneSummaryPub := messaging.NewZoneSummaryPublisher(bus)
	zoneCtrl := zone.NewController(
		api.ZoneID(zoneCfg.Zone.ZoneID),
		zoneRobots,
		cmdPub,
		zoneSummaryPub,
		bus,
		5*time.Second,
	)
	go func() {
		log.Printf("zone: %s (robots: %d)", zoneCfg.Zone.ZoneID, len(zoneRobots))
		_ = zoneCtrl.Run(ctx)
	}()

	// ---- Edge: one gateway per robot, or one Simulator for many robots ----
	const useSimulatorThreshold = 25
	if len(zoneRobots) >= useSimulatorThreshold {
		statusPub := messaging.NewRobotStatusPublisher(bus)
		sim := edge.NewSimulator(
			api.ZoneID(zoneCfg.Zone.ZoneID),
			zoneRobots,
			statusPub,
			bus,
			5*time.Second,
			2*time.Second,
		)
		go func() {
			log.Printf("edge: simulator for %d robots (zone %s)", len(zoneRobots), zoneCfg.Zone.ZoneID)
			_ = sim.Run(ctx)
		}()
	} else {
		for _, robotID := range zoneCfg.Zone.Robots {
			statusPub := messaging.NewRobotStatusPublisher(bus)
			gw := edge.NewGateway(
				api.RobotID(robotID),
				api.ZoneID(zoneCfg.Zone.ZoneID),
				"stub",
				statusPub,
				bus,
				2*time.Second,
				2*time.Second,
			)
			go func(id string) {
				log.Printf("edge: %s (zone %s)", id, zoneCfg.Zone.ZoneID)
				_ = gw.Run(ctx)
			}(robotID)
		}
	}

	<-ctx.Done()
	log.Println("shutting down...")
	_ = httpSrv.Shutdown(context.Background())
	log.Println("done")
}
