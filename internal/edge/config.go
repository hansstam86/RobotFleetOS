package edge

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds edge layer configuration (one edge process per robot or small cell).
type Config struct {
	Edge      EdgeConfig      `yaml:"edge"`
	Messaging MessagingConfig `yaml:"messaging"`
}

type EdgeConfig struct {
	RobotID  string `yaml:"robot_id"`
	ZoneID   string `yaml:"zone_id"`
	Protocol string `yaml:"robot_protocol"` // "stub", "opcua", "vendor-api", etc.
}

type MessagingConfig struct {
	Broker      string `yaml:"broker"`
	TopicPrefix string `yaml:"topic_prefix"`
}

// LoadConfig reads config from path. If path is empty, uses env EDGE_CONFIG or defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Edge: EdgeConfig{
			RobotID:  "robot-1",
			ZoneID:   "zone-1",
			Protocol: "stub",
		},
		Messaging: MessagingConfig{
			Broker:      "memory",
			TopicPrefix: "robotfleetos",
		},
	}
	if path == "" {
		if p := os.Getenv("EDGE_CONFIG"); p != "" {
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
	if cfg.Edge.RobotID == "" {
		cfg.Edge.RobotID = "robot-1"
	}
	if cfg.Edge.Protocol == "" {
		cfg.Edge.Protocol = "stub"
	}
	// Env override for deployment (e.g. 1000 edge pods with different ROBOT_ID/ZONE_ID).
	if id := os.Getenv("EDGE_ROBOT_ID"); id != "" {
		cfg.Edge.RobotID = id
	}
	if id := os.Getenv("EDGE_ZONE_ID"); id != "" {
		cfg.Edge.ZoneID = id
	}
	return cfg, nil
}
