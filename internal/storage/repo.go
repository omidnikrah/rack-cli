package storage

import "errors"

var ErrNotFound = errors.New("record not found")

type Identifiable[T any] interface {
	GetID() int
	SetID(id int) T
}

type Validator interface {
	Validate() error
}

type Record[T any] interface {
	Identifiable[T]
	Validator
}

type Repo[T any] interface {
	Load() ([]T, error)
	Get(id int) (T, error)
	Create(T) (T, error)
	Update(T) error
	Delete(id int) error
}
