package cache

import (
	"context"
	"fmt"
	"zadaniel0/internal/model"
	"zadaniel0/internal/storage"
)

type Cache struct {
	c map[string]model.Order

	pg *storage.Pg
}

func NewCache(pg *storage.Pg) *Cache {
	return &Cache{
		c:  make(map[string]model.Order),
		pg: pg,
	}
}

func (c *Cache) Set(key string, value model.Order) {
	c.c[key] = value
}

func (c *Cache) Get(ctx context.Context, key string) (model.Order, error) {
	v, ok := c.c[key]
	if !ok {

		order, err := c.pg.Get(ctx, key)
		if err != nil {
			return model.Order{}, fmt.Errorf("failed to pg get: %w", err)
		}

		c.Set(key, *order)

		return *order, nil
	}

	return v, nil
}

func (c *Cache) Setup(ctx context.Context) error {
	orders, err := c.pg.GetAllOrders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all orders: %w", err)
	}

	fmt.Printf("Loaded %d keys\n", len(orders))

	for _, order := range orders {
		c.Set(order.OrderUID, order)
	}

	return nil
}

func (c *Cache) GetAll() []model.Order {
	orders := make([]model.Order, 0, len(c.c))
	for _, order := range c.c {
		orders = append(orders, order)
	}
	return orders
}
