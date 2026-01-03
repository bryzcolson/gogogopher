package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	toml "github.com/pelletier/go-toml"
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

	address := fmt.Sprintf(":%d", config.Server.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("Failed to start server", "err", err)
		os.Exit(1)
	}
	slog.Info(fmt.Sprintf("Server listening on port %s", address))
	defer ln.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	connLimit := make(chan struct{}, config.Limits.MaxConnections)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				slog.Error("Error accepting connection", "err", err)
				continue
			}

			select {
			case connLimit <- struct{}{}:
				go func() {
					defer conn.Close()
					defer func() {
						<-connLimit
					}()

					slog.Info("Accepted connection", "remote_addr", conn.RemoteAddr().String())
					defer slog.Info("Closed connection", "remote_addr", conn.RemoteAddr().String())

					handleConn(conn)
				}()
			default:
				slog.Warn("Too many connections", "remote_addr", conn.RemoteAddr().String())
				conn.Close()
			}
		}
	}()

	sig := <-sigChan
	slog.Info("Received shutdown signal", "sig", sig.String())
}
