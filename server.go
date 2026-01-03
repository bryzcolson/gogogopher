package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func handleConn(conn net.Conn) {
	requestId := uuid.New().String()
	reqLogger := logger.With("requestId", requestId)

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		reqLogger.Error("Connection read error", "err", err)
		return
	}

	selector := strings.TrimSpace(line)
	reqLogger.Info("Received request for selector", "selector", selector)

	data, isGophermap, err := readPath(config.Server.HomeDir, selector)
	if err != nil {
		reqLogger.Error("Error reading path", "selector", selector, "err", err)
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		fmt.Fprintf(conn, "%s\r\n", line)
	}
	if isGophermap {
		fmt.Fprint(conn, ".\r\n")
	}
	reqLogger.Info("Completed request for selector", "selector", selector)
}

func readPath(basepath, selector string) ([]byte, bool, error) {
	path := filepath.Join(basepath, selector)
	info, err := os.Stat(path)
	if err != nil {
		return nil, false, errors.New("file not found")
	}

	isGophermap := false
	if info.IsDir() {
		path = filepath.Join(path, "gophermap")
		isGophermap = true
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, errors.New("file not found")
	}
	return data, isGophermap, nil
}
