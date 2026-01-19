package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	gopher "codeberg.org/bryzcolson/net-gopher"
	toml "github.com/pelletier/go-toml"
	"golang.org/x/net/netutil"
)

type Config struct {
	Server struct {
		Port    int    `toml:"port"`
		HomeDir string `toml:"homedir"`
	} `toml:"server"`
	Limits struct {
		MaxConnections int `toml:"max_connections"`
	} `toml:"limits"`
}

var (
	config Config
	logger *slog.Logger
)

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := toml.Unmarshal(data, &config); err != nil {
		return err
	}

	return nil
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	slog.Info("Initializing server...")
	defer slog.Info("Closing server")

	slog.Info("Loading config...")
	if err := loadConfig("config.toml"); err != nil {
		slog.Error("Failed to load config", "err", err)
		os.Exit(1)
	}
	slog.Info("Loaded config")

	gopher.HandleFunc("/", makeHandler(config.Server.HomeDir))

	addr := fmt.Sprintf(":%d", config.Server.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to listen", "addr", addr, "err", err)
		os.Exit(1)
	}

	if config.Limits.MaxConnections > 0 {
		ln = netutil.LimitListener(ln, config.Limits.MaxConnections)
		slog.Info("Connection limit enabled", "max", config.Limits.MaxConnections)
	}

	slog.Info("Server listening", "addr", addr)
	if err := gopher.Serve(ln, nil); err != nil {
		slog.Error("Server error", "err", err)
		os.Exit(1)
	}
}
