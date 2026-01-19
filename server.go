package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	gopher "codeberg.org/bryzcolson/net-gopher"
	"github.com/google/uuid"
)

func makeHandler(homedir string) func(gopher.ResponseWriter, *gopher.Request) {
	return func(w gopher.ResponseWriter, r *gopher.Request) {
		log := slog.With("requestId", uuid.New().String())
		log.Info("Request received", "selector", r.Selector)

		selector := strings.TrimPrefix(r.Selector, "/")
		if selector == "" {
			selector = "."
		}

		path := filepath.Join(homedir, selector)
		info, err := os.Stat(path)
		if err != nil {
			log.Error("Not found", "selector", selector)
			w.WriteError("Not found")
			return
		}

		if info.IsDir() {
			path = filepath.Join(homedir, selector, "gophermap")
		}

		data, err := os.ReadFile(path)
		if err != nil {
			log.Error("Cannot read file", "path", path, "err", err)
			w.WriteError("Cannot read file")
			return
		}

		if isBinary(data) {
			w.Write(data)
		} else {
			io.WriteString(w, string(data))
		}
		log.Info("Request completed", "selector", r.Selector)
	}
}

func isBinary(data []byte) bool {
	mime := http.DetectContentType(data)
	if strings.HasPrefix(mime, "text/") || strings.HasSuffix(mime, "+xml") || strings.HasSuffix(mime, "+json") {
		return false
	}

	switch mime {
	case "application/json",
		"application/xml",
		"application/javascript",
		"application/x-sh",
		"application/x-csh",
		"application/yaml",
		"application/x-yaml",
		"application/sql",
		"application/x-perl",
		"application/x-python",
		"application/x-ruby":
		return false
	}

	return true
}
