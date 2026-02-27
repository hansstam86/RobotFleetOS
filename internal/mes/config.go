package mes

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds MES configuration.
type Config struct {
	MES   MESConfig   `yaml:"mes"`
	Fleet FleetConfig `yaml:"fleet"`
}

// MESConfig is the MES server configuration.
type MESConfig struct {
	Listen string `yaml:"listen"`
}

// FleetConfig is the Fleet API client configuration.
type FleetConfig struct {
	APIURL string `yaml:"api_url"`
}

// LoadConfig reads config from path. If path is empty, uses env and defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		MES: MESConfig{
			Listen: ":8081",
		},
		Fleet: FleetConfig{
			APIURL: "http://localhost:8080",
		},
	}
	if path == "" {
		if p := os.Getenv("MES_CONFIG"); p != "" {
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
	if listen := os.Getenv("MES_LISTEN"); listen != "" {
		cfg.MES.Listen = listen
	}
	if cfg.MES.Listen == "" {
		cfg.MES.Listen = ":8081"
	}
	if cfg.Fleet.APIURL == "" {
		cfg.Fleet.APIURL = "http://localhost:8080"
	}
	return cfg, nil
}
