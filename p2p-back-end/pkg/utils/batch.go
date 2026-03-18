package utils

import "context"

func BatchSync[T any, K any](
	ctx context.Context,
	startID K,
	limit int,
	fetchFunc func(ctx context.Context, lastID K, limit int) ([]T, error),
	saveFunc func(ctx context.Context, data []T) error,
	idFunc func(item T) K,
) error {
	lastID := startID
	for {
		data, err := fetchFunc(ctx, lastID, limit)
		if err != nil {
			return err
		}

		if len(data) == 0 {
			break
		}

		if err := saveFunc(ctx, data); err != nil {
			return err
		}

		if len(data) < limit {
			break
		}

		lastID = idFunc(data[len(data)-1])
	}
	return nil
}
