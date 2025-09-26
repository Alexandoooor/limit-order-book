CREATE TABLE IF NOT EXISTS levels (
    side INTEGER NOT NULL,
    price INTEGER NOT NULL,
    volume INTEGER NOT NULL,
    count INTEGER NOT NULL,
    PRIMARY KEY(side, price)
);

CREATE TABLE IF NOT EXISTS level_orders (
    level_side TEXT NOT NULL,
    level_price INTEGER NOT NULL,
    order_id TEXT NOT NULL,
    PRIMARY KEY(level_side, level_price, order_id),
    FOREIGN KEY(level_side, level_price) REFERENCES levels(side, price),
    FOREIGN KEY(order_id) REFERENCES orders(id)
);

CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    side INTEGER NOT NULL,
    size INTEGER NOT NULL,
    remaining INTEGER NOT NULL,
    price INTEGER NOT NULL,
    time TEXT NOT NULL,
    next_id TEXT,
    prev_id TEXT,
    FOREIGN KEY(next_id) REFERENCES orders(id),
    FOREIGN KEY(prev_id) REFERENCES orders(id)
);

CREATE TABLE IF NOT EXISTS trades (
    id TEXT PRIMARY KEY,
    buy_order_id TEXT NOT NULL,
    sell_order_id TEXT NOT NULL,
    price INTEGER NOT NULL,
    size INTEGER NOT NULL,
    time TEXT NOT NULL,
    FOREIGN KEY(buy_order_id) REFERENCES orders(id),
    FOREIGN KEY(sell_order_id) REFERENCES orders(id)
);
