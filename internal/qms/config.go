package qms

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds QMS server configuration.
type Config struct {
	QMS QMSConfig `yaml:"qms"`
}

// QMSConfig is the server config.
type QMSConfig struct {
	Listen string `yaml:"listen"`
}

// LoadConfig reads config from path.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		QMS: QMSConfig{Listen: ":8084"},
	}
	if path == "" {
		if p := os.Getenv("QMS_CONFIG"); p != "" {
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
	if listen := os.Getenv("QMS_LISTEN"); listen != "" {
		cfg.QMS.Listen = listen
	}
	if cfg.QMS.Listen == "" {
		cfg.QMS.Listen = ":8084"
	}
	return cfg, nil
}
