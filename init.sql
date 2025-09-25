CREATE TABLE IF NOT EXISTS levels (
    rowid INTEGER PRIMARY KEY AUTOINCREMENT,
    side INTEGER NOT NULL,
    price INTEGER NOT NULL,
    volume INTEGER NOT NULL,
    count INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS level_orders (
    level_rowid INTEGER NOT NULL,
    order_id TEXT NOT NULL,
    PRIMARY KEY(level_rowid, order_id),
    FOREIGN KEY(level_rowid) REFERENCES levels(rowid),
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
