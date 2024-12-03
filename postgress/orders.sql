-- Таблица для структуры Order
CREATE TABLE orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255),
    entry VARCHAR(255),          
    locale VARCHAR(255),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(255),
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard VARCHAR(255)
);

-- Таблица для структуры Delivery
CREATE TABLE delivery (
    order_uid VARCHAR(255),
    name VARCHAR(255),
    phone VARCHAR(20),
    zip VARCHAR(20),
    city VARCHAR(255),
    address TEXT,
    region VARCHAR(255),
    email VARCHAR(255)
);

-- Таблица для структуры Payment
CREATE TABLE payment (
    order_uid VARCHAR(255),
    transaction VARCHAR(255),
    request_id VARCHAR(255),
    currency VARCHAR(20),
    provider VARCHAR(255),
    amount INT,
    payment_dt BIGINT,
    bank VARCHAR(255),
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

-- Таблица для структуры Item
CREATE TABLE items (
    order_uid VARCHAR(255),
    chrt_id INT,
    track_number VARCHAR(255),
    price INT,
    rid VARCHAR(255),
    name VARCHAR(255),
    sale INT,
    size VARCHAR(50),
    total_price INT,
    nm_id INT,
    brand VARCHAR(255),
    status INT
);
