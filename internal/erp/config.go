package erp

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ERP ERPConfig `yaml:"erp"`
	MES MESConfig `yaml:"mes"`
}

type ERPConfig struct {
	Listen string `yaml:"listen"`
}

type MESConfig struct {
	APIURL         string `yaml:"api_url"`
	DefaultAreaID  string `yaml:"default_area_id"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		ERP: ERPConfig{Listen: ":8087"},
		MES: MESConfig{APIURL: "http://localhost:8081", DefaultAreaID: "area-1"},
	}
	if path == "" {
		if p := os.Getenv("ERP_CONFIG"); p != "" {
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
	if listen := os.Getenv("ERP_LISTEN"); listen != "" {
		cfg.ERP.Listen = listen
	}
	if cfg.ERP.Listen == "" {
		cfg.ERP.Listen = ":8087"
	}
	if url := os.Getenv("MES_API_URL"); url != "" {
		cfg.MES.APIURL = url
	}
	if cfg.MES.APIURL == "" {
		cfg.MES.APIURL = "http://localhost:8081"
	}
	if a := os.Getenv("ERP_DEFAULT_AREA_ID"); a != "" {
		cfg.MES.DefaultAreaID = a
	}
	if cfg.MES.DefaultAreaID == "" {
		cfg.MES.DefaultAreaID = "area-1"
	}
	return cfg, nil
}
