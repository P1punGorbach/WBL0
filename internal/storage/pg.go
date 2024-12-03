package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"zadaniel0/internal/config"
	"zadaniel0/internal/model"

	"github.com/jackc/pgx/v5"
)

type Pg struct {
	datachan chan model.Order
	conn     *pgx.Conn
}

func NewPg(ctx context.Context, cfg config.Pg, ch chan model.Order) (*Pg, error) {
	conn, err := pgx.Connect(ctx, cfg.Url)
	if err != nil {
		return nil, fmt.Errorf("Fail to connect: %w", err)
	}
	return &Pg{datachan: ch, conn: conn}, nil
}
func (P *Pg) InsertNewOrder(ctx context.Context, order model.Order) error {
	_, err := P.conn.Exec(ctx,
		`INSERT INTO orders (
    order_uid, track_number, entry, locale,
    internal_signature, customer_id, delivery_service, shardkey, sm_id,
    date_created, oof_shard
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11
)`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert new order: %w", err)
	}
	return nil
}
func (P *Pg) InsertNewDelivery(ctx context.Context, id string, del model.Delivery) error {
	_, err := P.conn.Exec(ctx,
		`INSERT INTO delivery (
    order_uid, name, phone, zip, city, address, region, email
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)`,
		id,
		del.Name,
		del.Phone,
		del.Zip,
		del.City,
		del.Address,
		del.Region,
		del.Email,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert new delivery: %w", err)
	}
	return nil
}
func (P *Pg) InsertNewPayment(ctx context.Context, id string, pay model.Payment) error {
	_, err := P.conn.Exec(ctx,
		`INSERT INTO payment (
    order_uid, transaction, request_id, currency, provider, amount, payment_dt,
    bank, delivery_cost, goods_total, custom_fee
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11
)`,
		id,
		pay.Transaction,
		pay.RequestID,
		pay.Currency,
		pay.Provider,
		pay.Amount,
		pay.PaymentDT,
		pay.Bank,
		pay.DeliveryCost,
		pay.GoodsTotal,
		pay.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert new payment: %w", err)
	}
	return nil
}
func (P *Pg) InsertNewItem(ctx context.Context, id string, item model.Item) error {
	_, err := P.conn.Exec(ctx,
		`INSERT INTO items (
    order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price,
    nm_id, brand, status
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
	$12
)`,
		id,
		item.ChrtID,
		item.TrackNumber,
		item.Price,
		item.RID,
		item.Name,
		item.Sale,
		item.Size,
		item.TotalPrice,
		item.NmID,
		item.Brand,
		item.Status,
	)
	if err != nil {
		return fmt.Errorf("Failed to insert new payment: %w", err)
	}
	return nil
}
func (P *Pg) InsertOrder(ctx context.Context, order model.Order) error {
	err := P.InsertNewOrder(ctx, order)
	if err != nil {
		return fmt.Errorf("Failed to insert order: %w", err)
	}
	err = P.InsertNewDelivery(ctx, order.OrderUID, order.Delivery)
	if err != nil {
		return fmt.Errorf("Failed to insert delivery: %w", err)
	}
	err = P.InsertNewPayment(ctx, order.OrderUID, order.Payment)
	if err != nil {
		return fmt.Errorf("Failed to insert payment: %w", err)
	}
	for _, item := range order.Items {
		err = P.InsertNewItem(ctx, order.OrderUID, item)
		if err != nil {
			return fmt.Errorf("Failed to insert item: %w", err)
		}
	}
	return nil
}
func (P *Pg) Run(ctx context.Context) error {
	go func() {
		for {
			select {
			case message, ok := <-P.datachan:
				if !ok {
					log.Printf("message channel was closed")
				}
				err := P.InsertOrder(ctx, message)
				if err != nil {
					log.Printf("Failed to run: %v", err)
				}

			}
		}
	}()
	return nil
}

func (P *Pg) Close() {
	P.conn.Close(context.Background())
}

func (P *Pg) Get(ctx context.Context, id string) (*model.Order, error) {

	order, err := P.GetOrder(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get order: %w", err)
	}

	delivery, err := P.GetDelivery(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get delivery: %w", err)
	}

	order.Delivery = *delivery

	payment, err := P.GetPayment(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get payment: %w", err)
	}

	order.Payment = *payment

	items, err := P.GetItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get items: %w", err)
	}

	order.Items = items

	return order, nil
}

func (P *Pg) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	row := P.conn.QueryRow(ctx,
		`SELECT
		order_uid, track_number, entry, locale,
    internal_signature, customer_id, delivery_service, shardkey, sm_id,
    date_created, oof_shard FROM orders
		WHERE order_uid = $1`, id,
	)

	var order model.Order
	err := row.Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("Order not found")
		}

		return nil, err
	}

	return &order, nil
}

func (P *Pg) GetDelivery(ctx context.Context, id string) (*model.Delivery, error) {
	row := P.conn.QueryRow(ctx,
		`SELECT
    name, phone, zip, city, address, region, email
		FROM delivery
		WHERE order_uid = $1`,
		id,
	)

	var del model.Delivery
	err := row.Scan(
		&del.Name,
		&del.Phone,
		&del.Zip,
		&del.City,
		&del.Address,
		&del.Region,
		&del.Email,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("Order not found")
		}

		return nil, err
	}

	return &del, nil
}

func (P *Pg) GetPayment(ctx context.Context, id string) (*model.Payment, error) {
	row := P.conn.QueryRow(ctx,
		`SELECT
    transaction, request_id, currency, provider, amount, payment_dt,
    bank, delivery_cost, goods_total, custom_fee
		FROM payment
		WHERE order_uid = $1`,
		id,
	)

	var payment model.Payment
	err := row.Scan(
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDT,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("Order not found")
		}

		return nil, err
	}

	return &payment, nil
}

func (P *Pg) GetItems(ctx context.Context, id string) ([]model.Item, error) {
	rows, err := P.conn.Query(ctx,
		`SELECT
    chrt_id, track_number, price, rid, name, sale, size, total_price,
    nm_id, brand, status
		FROM items
		WHERE order_uid = $1`,
		id,
	)

	defer rows.Close()

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("Order not found")
		}

		return nil, err
	}

	var items []model.Item
	for rows.Next() {
		var item model.Item
		err = rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.RID,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)

		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (P *Pg) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	rows, err := P.conn.Query(ctx,
		`SELECT
    order_uid
		FROM orders`,
	)

	defer rows.Close()

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var orderUIDs []string
	for rows.Next() {
		var orderUID string
		err = rows.Scan(
			&orderUID,
		)

		if err != nil {
			return nil, err
		}

		orderUIDs = append(orderUIDs, orderUID)
	}

	// Get all orders

	var orders []model.Order
	for _, orderUID := range orderUIDs {
		order, err := P.Get(ctx, orderUID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, *order)
	}

	return orders, nil
}
