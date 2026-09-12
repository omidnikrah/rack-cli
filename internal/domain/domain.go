package domain

import (
	"rack-cli/internal/entity"
	"rack-cli/internal/storage"
)

type Storage = storage.Repo[entity.Domain]
