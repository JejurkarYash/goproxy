package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Port                       int      `json:"port"`
	BackendHealthCheckInterval int      `json:"backend_healthcheck_interval"`
	Backends                   []string `json:"backends"`
}

func LoadConfig() *Config {
	data, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}
