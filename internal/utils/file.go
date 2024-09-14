package utils

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func MarshalFromFile(filePath string, value any) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, value)
	return nil
}

func FilePaths(path string) ([]string, error) {
	var filePaths = make([]string, 0)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		filePaths = append(filePaths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return filePaths, err
}
