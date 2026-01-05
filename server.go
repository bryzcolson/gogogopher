package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidFilePath = errors.New("invalid file path")
	ErrFileNotFound    = errors.New("file not found")
)

type FileType int

const (
	FileTypeText = iota
	FileTypeGophermap
	FileTypeBinary
	FileTypeErr
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

	data, fileType, err := readPath(config.Server.HomeDir, selector)
	if err != nil {
		reqLogger.Error("Error reading path", "selector", selector, "err", err)
		return
	}

	if fileType == FileTypeBinary {
		_, err = conn.Write(data)
		if err != nil {
			reqLogger.Error("Error writing binary data", "err", err)
		}
	} else {
		for _, line := range strings.Split(string(data), "\n") {
			fmt.Fprintf(conn, "%s\r\n", line)
		}
		if fileType == FileTypeGophermap {
			fmt.Fprint(conn, ".\r\n")
		}
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
		return "", false, ErrInvalidFilePath
	}

	isDir := false
	if info.IsDir() {
		isDir = true
		selector = filepath.Join(selector, "gophermap")
	}
	return selector, isDir, nil
}

func readPath(basepath, selector string) ([]byte, FileType, error) {
	cleanSelector, isDir, err := cleanPath(basepath, selector)
	if err != nil {
		return nil, FileTypeErr, err
	}

	path := filepath.Join(basepath, cleanSelector)
	f, err := os.Open(path)
	if err != nil {
		return nil, FileTypeErr, ErrInvalidFilePath
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, FileTypeErr, ErrFileNotFound
	}

	var fileType FileType
	if isDir {
		fileType = FileTypeGophermap
	} else if isBinaryFile(data) {
		fileType = FileTypeBinary
	} else {
		fileType = FileTypeText
	}

	return data, fileType, nil
}

func isBinaryFile(data []byte) bool {
	mimeType := http.DetectContentType(data)
	if strings.HasPrefix(mimeType, "text/") {
		return false
	}

	textMimeTypes := map[string]bool{
		"application/json":       true,
		"application/xml":        true,
		"application/javascript": true,
		"application/x-sh":       true,
	}

	if textMimeTypes[mimeType] {
		return false
	}

	return true
}
