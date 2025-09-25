CREATE TABLE levels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    side TEXT NOT NULL,
    price INTEGER NOT NULL,
    volume INTEGER NOT NULL,
    count INTEGER NOT NULL
);

CREATE TABLE level_orders (
    level_id INTEGER NOT NULL,
    order_id TEXT NOT NULL,
    FOREIGN KEY(level_id) REFERENCES levels(id),
    FOREIGN KEY(order_id) REFERENCES orders(id),
    PRIMARY KEY (level_id, order_id)
);

CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    side TEXT NOT NULL,
    size INTEGER NOT NULL,
    remaining INTEGER NOT NULL,
    price INTEGER NOT NULL,
    time DATETIME NOT NULL,
    next_id TEXT,
    prev_id TEXT,
    FOREIGN KEY(next_id) REFERENCES orders(id),
    FOREIGN KEY(prev_id) REFERENCES orders(id)
);

CREATE TABLE trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    buy_order_id TEXT NOT NULL,
    sell_order_id TEXT NOT NULL,
    price INTEGER NOT NULL,
    size INTEGER NOT NULL,
    time DATETIME NOT NULL,
    FOREIGN KEY(buy_order_id) REFERENCES orders(id),
    FOREIGN KEY(sell_order_id) REFERENCES orders(id)
);
