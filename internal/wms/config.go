package wms

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds WMS configuration.
type Config struct {
	WMS WMSConfig `yaml:"wms"`
	Fleet FleetConfig `yaml:"fleet"`
}

// WMSConfig is the WMS server configuration.
type WMSConfig struct {
	Listen           string `yaml:"listen"`
	WarehouseAreaID  string `yaml:"warehouse_area_id"`
}

// FleetConfig is the Fleet API client configuration.
type FleetConfig struct {
	APIURL string `yaml:"api_url"`
}

// LoadConfig reads config from path. Uses env and defaults if path empty.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		WMS: WMSConfig{
			Listen:          ":8082",
			WarehouseAreaID: "area-1",
		},
		Fleet: FleetConfig{
			APIURL: "http://localhost:8080",
		},
	}
	if path == "" {
		if p := os.Getenv("WMS_CONFIG"); p != "" {
			path = p
		}
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}
	if url := os.Getenv("FLEET_API_URL"); url != "" {
		cfg.Fleet.APIURL = url
	}
	if listen := os.Getenv("WMS_LISTEN"); listen != "" {
		cfg.WMS.Listen = listen
	}
	if cfg.WMS.Listen == "" {
		cfg.WMS.Listen = ":8082"
	}
	if cfg.WMS.WarehouseAreaID == "" {
		cfg.WMS.WarehouseAreaID = "area-1"
	}
	if cfg.Fleet.APIURL == "" {
		cfg.Fleet.APIURL = "http://localhost:8080"
	}
	return cfg, nil
}
