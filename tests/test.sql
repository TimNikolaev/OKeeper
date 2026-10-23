INSERT INTO users (username, password_hash, full_name, role) VALUES
('waiter1', 'hash1', 'Иван Официантов', 'waiter'),
('waiter2', 'hash2', 'Пётр Официантов', 'waiter'),
('cook1',   'hash3', 'Мария Поварова',   'cook'),
('admin1',  'hash4', 'Сидор Админов',    'admin');

INSERT INTO dishes (category, name, description, price, is_available) VALUES
('напитки',  'Кофе',        'Американо',       350, 1),
('напитки',  'Чай',         'Чёрный',          250, 1),
('горячее',  'Паста',       'Карбонара',       890, 1),
('горячее',  'Стейк',       'Стейк из говядины', 1200, 1),
('десерты',  'Чизкейк',     'Нью-Йорк',        450, 1);

INSERT INTO orders (table_num, waiter_id, status, comment) VALUES
(1, 1, 'new',      NULL),
(2, 1, 'cooking',  'Без лука'),
(3, 2, 'ready',    NULL),
(4, 2, 'served',   'Гость торопится');

INSERT INTO orders_dishes (order_id, dish_id, quantity, price, comment) VALUES
(1, 1, 2, 350, 'Без сахара'),
(1, 3, 1, 890, NULL),
(2, 2, 1, 250, NULL),
(2, 4, 1, 1200, 'Средняя прожарка'),
(3, 1, 1, 350, NULL),
(3, 5, 2, 450, NULL),
(4, 2, 2, 250, NULL);
