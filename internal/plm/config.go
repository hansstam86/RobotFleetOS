package plm

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds PLM server configuration.
type Config struct {
	PLM PLMConfig `yaml:"plm"`
}

// PLMConfig is the server config.
type PLMConfig struct {
	Listen string `yaml:"listen"`
}

// LoadConfig reads config from path.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		PLM: PLMConfig{Listen: ":8086"},
	}
	if path == "" {
		if p := os.Getenv("PLM_CONFIG"); p != "" {
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
	if listen := os.Getenv("PLM_LISTEN"); listen != "" {
		cfg.PLM.Listen = listen
	}
	if cfg.PLM.Listen == "" {
		cfg.PLM.Listen = ":8086"
	}
	return cfg, nil
}
