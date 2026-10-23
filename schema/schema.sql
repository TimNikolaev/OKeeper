CREATE TABLE IF NOT EXISTS users
(
    id            INTEGER PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name     TEXT NOT NULL,
    role          TEXT NOT NULL CHECK (role IN ('waiter', 'cook', 'admin')),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders
(
    id         INTEGER PRIMARY KEY,
    table_num  INTEGER NOT NULL,
    waiter_id  INTEGER NOT NULL REFERENCES users(id),
    status     TEXT NOT NULL DEFAULT 'new'
               CHECK (status IN ('new','cooking','ready','served','paid','cancelled')),
    comment    TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dishes
(
    id           INTEGER PRIMARY KEY,
    category     TEXT NOT NULL,
    name         TEXT NOT NULL,
    description  TEXT,
    price        INTEGER NOT NULL CHECK (price >= 0),
    is_available INTEGER NOT NULL DEFAULT 1 CHECK (is_available IN (0, 1)),
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders_dishes
(
    id           INTEGER PRIMARY KEY,
    order_id     INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    dish_id      INTEGER NOT NULL REFERENCES dishes(id) ON DELETE RESTRICT,
    quantity     INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    price        INTEGER NOT NULL CHECK (price >= 0),
    comment      TEXT
);

CREATE INDEX IF NOT EXISTS idx_orders_dishes_order ON orders_dishes(order_id);
CREATE INDEX IF NOT EXISTS idx_orders_dishes_dish  ON orders_dishes(dish_id);
