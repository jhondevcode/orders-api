package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	api_errors "github.com/jhondevcode/orders-api/errors"
	"github.com/jhondevcode/orders-api/model"
	"github.com/jhondevcode/orders-api/types"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Client *redis.Client
}

func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepository) Insert(ctx context.Context, order model.Order) error {
	if data, err := json.Marshal(order); err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	} else {
		key := orderIDKey(order.OrderID)

		txn := r.Client.TxPipeline()

		res := r.Client.SetNX(ctx, key, string(data), 0)
		if err := res.Err(); err != nil {
			txn.Discard()
			return fmt.Errorf("failed to set: %w", err)
		}

		if err := r.Client.SAdd(ctx, "orders", key).Err(); err != nil {
			txn.Discard()
			return fmt.Errorf("failed to add to orders set: %w", err)
		}

		if _, err := txn.Exec(ctx); err != nil {
			return fmt.Errorf("failed to exec: %w", err)
		}
	}
	return nil
}

func (r *RedisRepository) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey((id))
	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, api_errors.ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get order: %w", err)
	}

	var order model.Order
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}
	return order, nil
}

func (r *RedisRepository) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return api_errors.ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get order: %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove from orders set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepository) Update(ctx context.Context, order model.Order) error {
	if data, err := json.Marshal(order); err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	} else {
		key := orderIDKey(order.OrderID)
		err := r.Client.SetXX(ctx, key, string(data), 0).Err()
		if errors.Is(err, redis.Nil) {
			return api_errors.ErrNotExist
		} else if err != nil {
			return fmt.Errorf("set ordder: %w", err)
		}
	}

	return nil
}

func (r *RedisRepository) FindAll(ctx context.Context, page types.FindAllPage) (types.FindResult, error) {
	res := r.Client.SScan(ctx, "orders", uint64(page.Offset), "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return types.FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	if len(keys) == 0 {
		return types.FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return types.FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]model.Order, len(xs))
	for i, x := range xs {
		x := x.(string)
		var order model.Order

		if err := json.Unmarshal([]byte(x), &order); err != nil {
			return types.FindResult{}, fmt.Errorf("failed to decode order json: %w", err)
		}

		orders[i] = order
	}

	return types.FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
