package area

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds area layer configuration.
type Config struct {
	Area      AreaConfig      `yaml:"area"`
	Messaging MessagingConfig `yaml:"messaging"`
}

type AreaConfig struct {
	AreaID string   `yaml:"area_id"`
	Zones  []string `yaml:"zones"` // zone IDs this area owns (for dispatching work)
}

type MessagingConfig struct {
	Broker      string `yaml:"broker"`
	TopicPrefix string `yaml:"topic_prefix"`
}

// LoadConfig reads config from path. If path is empty, uses env AREA_CONFIG or defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Area: AreaConfig{
			AreaID: "area-1",
			Zones:  []string{"zone-1"},
		},
		Messaging: MessagingConfig{
			Broker:      "memory",
			TopicPrefix: "robotfleetos",
		},
	}
	if path == "" {
		if p := os.Getenv("AREA_CONFIG"); p != "" {
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
	if cfg.Area.AreaID == "" {
		cfg.Area.AreaID = "area-1"
	}
	if len(cfg.Area.Zones) == 0 {
		cfg.Area.Zones = []string{"zone-1"}
	}
	return cfg, nil
}
