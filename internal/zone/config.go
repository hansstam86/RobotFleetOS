package zone

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds zone layer configuration.
type Config struct {
	Zone      ZoneConfig      `yaml:"zone"`
	Messaging MessagingConfig `yaml:"messaging"`
}

type ZoneConfig struct {
	ZoneID  string   `yaml:"zone_id"`
	AreaID  string   `yaml:"area_id"`
	Robots  []string `yaml:"robots"` // robot IDs in this zone (for dispatching commands)
}

type MessagingConfig struct {
	Broker      string `yaml:"broker"`
	TopicPrefix string `yaml:"topic_prefix"`
}

// LoadConfig reads config from path. If path is empty, uses env ZONE_CONFIG or defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Zone: ZoneConfig{
			ZoneID: "zone-1",
			AreaID: "area-1",
			Robots: []string{"robot-1"},
		},
		Messaging: MessagingConfig{
			Broker:      "memory",
			TopicPrefix: "robotfleetos",
		},
	}
	if path == "" {
		if p := os.Getenv("ZONE_CONFIG"); p != "" {
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
	if cfg.Zone.ZoneID == "" {
		cfg.Zone.ZoneID = "zone-1"
	}
	if len(cfg.Zone.Robots) == 0 {
		cfg.Zone.Robots = []string{"robot-1"}
	}
	return cfg, nil
}
