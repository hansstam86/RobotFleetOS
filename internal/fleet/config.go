package fleet

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds fleet layer configuration.
type Config struct {
	Fleet     FleetConfig     `yaml:"fleet"`
	Messaging MessagingConfig `yaml:"messaging"`
	State     StateConfig     `yaml:"state"`
}

type FleetConfig struct {
	SchedulerWorkers int    `yaml:"scheduler_workers"`
	APIListen        string `yaml:"api_listen"`
}

type MessagingConfig struct {
	Broker      string `yaml:"broker"`
	TopicPrefix string `yaml:"topic_prefix"`
}

type StateConfig struct {
	Endpoints []string `yaml:"endpoints"`
}

// LoadConfig reads config from path. If path is empty, uses default in-memory/dev settings.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Fleet: FleetConfig{
			SchedulerWorkers: 10,
			APIListen:        ":8080",
		},
		Messaging: MessagingConfig{
			Broker:      "memory",
			TopicPrefix: "robotfleetos",
		},
		State: StateConfig{Endpoints: []string{"memory"}},
	}
	if path == "" {
		if p := os.Getenv("FLEET_CONFIG"); p != "" {
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
	if cfg.Fleet.APIListen == "" {
		cfg.Fleet.APIListen = ":8080"
	}
	return cfg, nil
}
