package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	WorkerPool struct {
		NumWorkers    int `yaml:"num_workers"`
		TaskQueueSize int `yaml:"task_queue_size"`
	} `yaml:"workerpool"`
	Storage struct {
		DownloadsDir string `yaml:"downloads_dir"`
	} `yaml:"storage"`
	Tasks struct {
		File string `yaml:"file"`
	} `yaml:"tasks"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
