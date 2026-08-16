package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Identifiable[T any] interface {
	GetID() int
	SetID(id int) T
}

type JSONStorage[T Identifiable[T]] struct {
	DefaultStoragePath string
	items              []T
	nextID             int
	loaded             bool
}

const (
	serversFile     = "servers.json"
	tempServersFile = "temp-servers.json"
)

func New[T Identifiable[T]](defaultPath string) *JSONStorage[T] {
	return &JSONStorage[T]{DefaultStoragePath: defaultPath}
}

func (s *JSONStorage[T]) Load() ([]T, error) {
	file, err := os.ReadFile(filepath.Join(s.DefaultStoragePath, serversFile))
	if err != nil {
		if os.IsNotExist(err) {
			s.items = []T{}
			s.nextID = 1
			s.loaded = true
			return []T{}, nil
		}
		return nil, err
	}

	if len(file) == 0 {
		s.items = []T{}
		s.nextID = 1
		s.loaded = true
		return []T{}, nil
	}

	var items []T
	if err := json.Unmarshal(file, &items); err != nil {
		return nil, err
	}

	s.items = items

	id := 1
	for index := range items {
		if items[index].GetID() >= id {
			id = items[index].GetID() + 1
		}
	}

	s.nextID = id
	s.loaded = true

	return items, nil
}

func (s *JSONStorage[T]) Save(item T) error {
	if !s.loaded {
		_, err := s.Load()
		if err != nil {
			return err
		}
	}

	item = item.SetID(s.nextID)

	s.items = append(s.items, item)
	s.nextID += 1

	data, err := json.Marshal(s.items)
	if err != nil {
		return err
	}

	tempFilePath := filepath.Join(s.DefaultStoragePath, tempServersFile)

	wErr := os.WriteFile(tempFilePath, data, 0644)
	if wErr != nil {
		return wErr
	}

	return os.Rename(tempFilePath, filepath.Join(s.DefaultStoragePath, serversFile))
}
