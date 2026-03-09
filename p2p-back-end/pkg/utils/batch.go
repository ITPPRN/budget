package utils

func BatchSync[T any, K any](
	startID K,
	limit int,
	fetchFunc func(lastID K, limit int) ([]T, error),
	saveFunc func(data []T) error,
	idFunc func(item T) K,
) error {
	lastID := startID
	for {
		data, err := fetchFunc(lastID, limit)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			break
		}

		if err := saveFunc(data); err != nil {
			return err
		}

		if len(data) < limit {
			break
		}

		lastID = idFunc(data[len(data)-1])
	}
	return nil
}
