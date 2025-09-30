-- Levels table
CREATE TABLE IF NOT EXISTS levels (
    side INTEGER NOT NULL,
    price INTEGER NOT NULL,
    volume INTEGER NOT NULL,
    count INTEGER NOT NULL,
    PRIMARY KEY (side, price)
);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    side INTEGER NOT NULL,
    size INTEGER NOT NULL,
    remaining INTEGER NOT NULL,
    price INTEGER NOT NULL,
    time TIMESTAMP NOT NULL,
    next_id TEXT,
    prev_id TEXT,
    CONSTRAINT fk_next FOREIGN KEY (next_id) REFERENCES orders(id),
    CONSTRAINT fk_prev FOREIGN KEY (prev_id) REFERENCES orders(id)
);

-- Level_orders table
CREATE TABLE IF NOT EXISTS level_orders (
    level_side INTEGER NOT NULL,
    level_price INTEGER NOT NULL,
    order_id TEXT NOT NULL,
    PRIMARY KEY(level_side, level_price, order_id),
    CONSTRAINT fk_level FOREIGN KEY(level_side, level_price) REFERENCES levels(side, price),
    CONSTRAINT fk_order FOREIGN KEY(order_id) REFERENCES orders(id)
);

-- Trades table
CREATE TABLE IF NOT EXISTS trades (
    id TEXT PRIMARY KEY,
    buy_order_id TEXT NOT NULL,
    sell_order_id TEXT NOT NULL,
    price INTEGER NOT NULL,
    size INTEGER NOT NULL,
    time TIMESTAMP NOT NULL,
    CONSTRAINT fk_buy FOREIGN KEY (buy_order_id) REFERENCES orders(id),
    CONSTRAINT fk_sell FOREIGN KEY (sell_order_id) REFERENCES orders(id)
);

-- Triggers in PostgreSQL require functions
-- Trigger function for level_orders AFTER INSERT
CREATE OR REPLACE FUNCTION level_orders_after_insert() RETURNS TRIGGER AS $$
BEGIN
    UPDATE levels
    SET
        count = count + 1,
        volume = volume + (SELECT remaining FROM orders WHERE id = NEW.order_id)
    WHERE side = NEW.level_side AND price = NEW.level_price;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER level_orders_after_insert
AFTER INSERT ON level_orders
FOR EACH ROW
EXECUTE FUNCTION level_orders_after_insert();

-- Trigger function for level_orders AFTER DELETE
CREATE OR REPLACE FUNCTION level_orders_after_delete() RETURNS TRIGGER AS $$
BEGIN
    UPDATE levels
    SET
        count = count - 1,
        volume = volume - (SELECT remaining FROM orders WHERE id = OLD.order_id)
    WHERE side = OLD.level_side AND price = OLD.level_price;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER level_orders_after_delete
AFTER DELETE ON level_orders
FOR EACH ROW
EXECUTE FUNCTION level_orders_after_delete();

-- Trigger function for orders AFTER UPDATE OF remaining
CREATE OR REPLACE FUNCTION orders_after_update_remaining() RETURNS TRIGGER AS $$
BEGIN
    UPDATE levels
    SET volume = volume + (NEW.remaining - OLD.remaining)
    WHERE side = NEW.side AND price = NEW.price;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER orders_after_update_remaining
AFTER UPDATE OF remaining ON orders
FOR EACH ROW
EXECUTE FUNCTION orders_after_update_remaining();
