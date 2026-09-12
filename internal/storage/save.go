package storage

func Save[T Record[T]](repo Repo[T], v T) (T, error) {
	if err := v.Validate(); err != nil {
		var zero T
		return zero, err
	}

	if v.GetID() == 0 {
		return repo.Create(v)
	}

	return v, repo.Update(v)
}
