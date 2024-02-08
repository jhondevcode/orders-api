package types

import "github.com/jhondevcode/orders-api/model"

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}
