package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/thalesraymond/web-crawler-go/internal"
)

type FileStorage struct {
	dataPath string
	mu       sync.Mutex
}

func NewFileStorage(dataPath string) *FileStorage {
	return &FileStorage{
		dataPath: dataPath,
	}
}

func (s *FileStorage) Save(result *internal.CrawlResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if result == nil {
		return errors.New("crawl result cannot be nil")
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")

	if err != nil {
		return err
	}

	encodedUrl := base64.URLEncoding.EncodeToString([]byte(result.URL))

	filePath := filepath.Join(s.dataPath, encodedUrl+".json")

	const readWriteOwner = os.FileMode(0400) | os.FileMode(0200)
	const readGroup = os.FileMode(0040)
	const readOthers = os.FileMode(0004)

	var perm644 os.FileMode = readWriteOwner | readGroup | readOthers

	if err := os.WriteFile(filePath, jsonData, perm644); err != nil {
		return err
	}

	return nil

}
