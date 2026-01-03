package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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

func cleanPath(basepath, selector string) (string, bool, error) {
	root, err := os.OpenRoot(basepath)
	if err != nil {
		return "", false, err
	}
	defer root.Close()

	selector = strings.TrimPrefix(selector, "/")
	selector = filepath.Clean(selector)

	info, err := root.Stat(selector)
	if err != nil {
		return "", false, errors.New("file not found")
	}

	isDir := false
	if info.IsDir() {
		isDir = true
		selector = filepath.Join(selector, "gophermap")
	}
	return selector, isDir, nil
}

func readPath(basepath, selector string) ([]byte, bool, error) {
	cleanSelector, isDir, err := cleanPath(basepath, selector)
	if err != nil {
		return nil, isDir, err
	}

	path := filepath.Join(basepath, cleanSelector)
	f, err := os.Open(path)
	if err != nil {
		return nil, isDir, errors.New("invalid file path")
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, isDir, errors.New("file not found")
	}
	return data, isDir, nil
}
