package main

import (
	"fmt"
	"os"
	"strings"
)

func listFiles(path string) ([]byte, error) {
	entries, err := os.ReadDir(strings.TrimSpace(path))
	if err != nil {
		Println("Error reading directory:", err)
		return nil, err
	}

	result := "\n"
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			Println("Error getting file info:", err)
			continue
		}

		result += fmt.Sprintf("%s\t%s\n", info.Mode().Perm(), entry.Name())
	}

	return []byte(result), nil
}
