package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr string
	DBPath     string
	DebugFlag  bool
}

const DBAddr = "postgres://postgres:6406655@localhost:5432/postgres?sslmode=disable"

func ReadConfig() *Config {
	var cfg Config
	var debugEnable *bool
	flag.StringVar(&cfg.ServerAddr, "a", ":8080", "server address")
	flag.StringVar(&cfg.DBPath, "d", DBAddr, "path to sqlite db")
	debugEnable = flag.Bool("debug", false, "debug on")
	flag.Parse()

	cfg.DebugFlag = *debugEnable

	if sAddr := os.Getenv("SERVER_ADDR"); sAddr != "" {
		cfg.ServerAddr = sAddr
	}
	if dbPath := os.Getenv("DATA_BASE_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}

	return &cfg
}
