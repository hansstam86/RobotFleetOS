package traceability

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds traceability server configuration.
type Config struct {
	Traceability TraceConfig `yaml:"traceability"`
}

// TraceConfig is the server config.
type TraceConfig struct {
	Listen string `yaml:"listen"`
}

// LoadConfig reads config from path. Uses env and defaults if path empty.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Traceability: TraceConfig{
			Listen: ":8083",
		},
	}
	if path == "" {
		if p := os.Getenv("TRACEABILITY_CONFIG"); p != "" {
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
	if listen := os.Getenv("TRACEABILITY_LISTEN"); listen != "" {
		cfg.Traceability.Listen = listen
	}
	if cfg.Traceability.Listen == "" {
		cfg.Traceability.Listen = ":8083"
	}
	return cfg, nil
}
