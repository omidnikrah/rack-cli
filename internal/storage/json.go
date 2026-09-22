package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

type JSONRepo[T Identifiable[T]] struct {
	path string

	mu     sync.Mutex
	items  []T
	nextID int
	loaded bool
}

type jsonFile[T any] struct {
	NextID int `json:"next_id"`
	Items  []T `json:"items"`
}

func NewJSONRepo[T Identifiable[T]](path string) *JSONRepo[T] {
	return &JSONRepo[T]{path: path}
}

func (r *JSONRepo[T]) Load() ([]T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureLoaded(); err != nil {
		return nil, err
	}

	return slices.Clone(r.items), nil
}

func (r *JSONRepo[T]) Get(id int) (T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var zero T
	if err := r.ensureLoaded(); err != nil {
		return zero, err
	}

	i := r.indexOf(id)
	if i < 0 {
		return zero, fmt.Errorf("get id %d: %w", id, ErrNotFound)
	}

	return r.items[i], nil
}

func (r *JSONRepo[T]) Create(item T) (T, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var zero T
	if err := r.ensureLoaded(); err != nil {
		return zero, err
	}

	item = item.SetID(r.nextID)
	items := append(slices.Clone(r.items), item)
	if err := r.write(items, r.nextID+1); err != nil {
		return zero, err
	}

	r.items, r.nextID = items, r.nextID+1
	return item, nil
}

func (r *JSONRepo[T]) Update(item T) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureLoaded(); err != nil {
		return err
	}

	i := r.indexOf(item.GetID())
	if i < 0 {
		return fmt.Errorf("update id %d: %w", item.GetID(), ErrNotFound)
	}

	items := slices.Clone(r.items)
	items[i] = item
	if err := r.write(items, r.nextID); err != nil {
		return err
	}

	r.items = items
	return nil
}

func (r *JSONRepo[T]) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureLoaded(); err != nil {
		return err
	}

	i := r.indexOf(id)
	if i < 0 {
		return fmt.Errorf("delete id %d: %w", id, ErrNotFound)
	}

	items := slices.Delete(slices.Clone(r.items), i, i+1)
	if err := r.write(items, r.nextID); err != nil {
		return err
	}

	r.items = items
	return nil
}

func (r *JSONRepo[T]) ensureLoaded() error {
	if r.loaded {
		return nil
	}

	raw, err := os.ReadFile(r.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// first run: start with an empty list
	case err != nil:
		return err
	case len(raw) > 0:
		var f jsonFile[T]
		if err := json.Unmarshal(raw, &f); err != nil {
			return fmt.Errorf("parsing %s: %w", r.path, err)
		}
		r.items, r.nextID = f.Items, f.NextID
	}

	r.nextID = max(r.nextID, r.maxID()+1)
	r.loaded = true
	return nil
}

func (r *JSONRepo[T]) indexOf(id int) int {
	return slices.IndexFunc(r.items, func(it T) bool { return it.GetID() == id })
}

func (r *JSONRepo[T]) maxID() int {
	highest := 0
	for _, it := range r.items {
		highest = max(highest, it.GetID())
	}
	return highest
}

func (r *JSONRepo[T]) write(items []T, nextID int) error {
	data, err := json.MarshalIndent(jsonFile[T]{NextID: nextID, Items: items}, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(r.path), filepath.Base(r.path)+".tmp-*")
	if err != nil {
		return err
	}

	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	if _, err := tmp.Write(data); err != nil {
		return err
	}

	if err := tmp.Chmod(0o600); err != nil {
		return err
	}

	if err := tmp.Sync(); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), r.path)
}

func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "rack-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}
