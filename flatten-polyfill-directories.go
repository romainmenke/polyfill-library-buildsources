package main

import (
	"os"
	"path/filepath"
	"strings"
)

func flattenPolyfillDirectories(directory string) ([]string, error) {
	var dirs []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(info.Name(), "__") {
			return nil
		}

		if info.IsDir() {
			dirs = append(dirs, strings.TrimPrefix(strings.TrimPrefix(path, filepath.Base(directory)), "/"))
		}

		return nil
	})

	return dirs, err
}
