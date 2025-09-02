package main

import (
	"errors"
	"os"
	"strings"
	"unicode/utf8"
)


const maxFileSize = 10 * 1024 * 1024 // 10 MB


func isValidUTF8(data []byte) bool {
    return utf8.Valid(data)
}

func cat(path string) ([]byte, error) {
	// Open the file
	file, err := os.Open(strings.TrimSpace(path))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file info to check size
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if info.Size() > maxFileSize {
		return nil, errors.New("file too large (over 10 MB) Use download instead")
	}

	// Read file content
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return nil, err
	}

	if !isValidUTF8(data) {
        return nil, errors.New("file contains invalid UTF‑8 — use download instead")
    }

	return data, nil
}
