package cmms

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds CMMS server configuration.
type Config struct {
	CMMS  CMMSConfig  `yaml:"cmms"`
	Fleet FleetConfig `yaml:"fleet"`
}

// CMMSConfig is the server config.
type CMMSConfig struct {
	Listen string `yaml:"listen"`
}

// FleetConfig is the Fleet API client config.
type FleetConfig struct {
	APIURL string `yaml:"api_url"`
}

// LoadConfig reads config from path. Uses env and defaults if path empty.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		CMMS: CMMSConfig{Listen: ":8085"},
		Fleet: FleetConfig{APIURL: "http://localhost:8080"},
	}
	if path == "" {
		if p := os.Getenv("CMMS_CONFIG"); p != "" {
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
	if listen := os.Getenv("CMMS_LISTEN"); listen != "" {
		cfg.CMMS.Listen = listen
	}
	if cfg.CMMS.Listen == "" {
		cfg.CMMS.Listen = ":8085"
	}
	if url := os.Getenv("FLEET_API_URL"); url != "" {
		cfg.Fleet.APIURL = url
	}
	if cfg.Fleet.APIURL == "" {
		cfg.Fleet.APIURL = "http://localhost:8080"
	}
	return cfg, nil
}
