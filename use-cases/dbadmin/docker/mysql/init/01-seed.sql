-- Seed data for MySQL test database
-- This runs automatically on first container start

USE testdb;

-- Users table with various data types
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    age INT,
    balance DECIMAL(10, 2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    metadata JSON,
    INDEX idx_email (email),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Products table for testing relationships
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INT DEFAULT 0,
    category ENUM('electronics', 'clothing', 'books', 'food', 'other') DEFAULT 'other',
    tags SET('new', 'sale', 'featured', 'clearance'),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_price (price)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Orders table for testing foreign keys
CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    total DECIMAL(10, 2) NOT NULL,
    status ENUM('pending', 'processing', 'shipped', 'delivered', 'cancelled') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Order items for testing joins
CREATE TABLE IF NOT EXISTS order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    INDEX idx_order (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample users
INSERT INTO users (username, email, age, balance, is_active, metadata) VALUES
('alice', 'alice@example.com', 28, 1500.50, TRUE, '{"preferences": {"theme": "dark", "notifications": true}}'),
('bob', 'bob@example.com', 34, 2300.75, TRUE, '{"preferences": {"theme": "light", "notifications": false}}'),
('charlie', 'charlie@example.com', 42, 890.25, TRUE, '{"preferences": {"theme": "auto", "notifications": true}}'),
('diana', 'diana@example.com', 29, 3200.00, FALSE, '{"preferences": {"theme": "dark", "notifications": true}}'),
('eve', 'eve@example.com', 31, 1750.80, TRUE, '{"preferences": {"theme": "light", "notifications": false}}'),
('frank', 'frank@example.com', 45, 4100.60, TRUE, '{"preferences": {"theme": "dark", "notifications": true}}'),
('grace', 'grace@example.com', 27, 950.30, TRUE, '{"preferences": {"theme": "light", "notifications": true}}'),
('henry', 'henry@example.com', 38, 2800.45, FALSE, '{"preferences": {"theme": "auto", "notifications": false}}'),
('iris', 'iris@example.com', 33, 1920.15, TRUE, '{"preferences": {"theme": "dark", "notifications": true}}'),
('jack', 'jack@example.com', 26, 1100.90, TRUE, '{"preferences": {"theme": "light", "notifications": true}}');

-- Insert sample products
INSERT INTO products (name, description, price, stock, category, tags) VALUES
('Laptop Pro', 'High-performance laptop with 16GB RAM', 1299.99, 50, 'electronics', 'new,featured'),
('Wireless Mouse', 'Ergonomic wireless mouse', 29.99, 200, 'electronics', 'sale'),
('Programming Book', 'Learn Go Programming', 49.99, 100, 'books', 'featured'),
('T-Shirt', 'Cotton t-shirt, various sizes', 19.99, 500, 'clothing', 'new,sale'),
('Coffee Beans', 'Premium arabica coffee beans 1kg', 24.99, 150, 'food', 'featured'),
('Mechanical Keyboard', 'RGB mechanical keyboard', 89.99, 75, 'electronics', 'new'),
('Novel', 'Bestselling fiction novel', 14.99, 300, 'books', 'clearance'),
('Jeans', 'Classic blue jeans', 59.99, 250, 'clothing', 'featured'),
('Chocolate Bar', 'Dark chocolate 100g', 3.99, 1000, 'food', 'sale'),
('Headphones', 'Noise-cancelling headphones', 199.99, 60, 'electronics', 'new,featured');

-- Insert sample orders
INSERT INTO orders (user_id, total, status) VALUES
(1, 1329.98, 'delivered'),
(2, 79.98, 'shipped'),
(3, 124.97, 'processing'),
(4, 29.99, 'pending'),
(5, 249.98, 'delivered'),
(1, 49.99, 'shipped'),
(6, 89.99, 'processing'),
(7, 34.98, 'pending'),
(8, 199.99, 'cancelled'),
(9, 64.98, 'delivered');

-- Insert sample order items
INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
(1, 1, 1, 1299.99),
(1, 2, 1, 29.99),
(2, 3, 1, 49.99),
(2, 2, 1, 29.99),
(3, 4, 2, 19.99),
(3, 5, 1, 24.99),
(3, 9, 5, 3.99),
(4, 2, 1, 29.99),
(5, 10, 1, 199.99),
(5, 3, 1, 49.99),
(6, 3, 1, 49.99),
(7, 6, 1, 89.99),
(8, 4, 1, 19.99),
(8, 9, 2, 3.99),
(9, 10, 1, 199.99),
(10, 7, 1, 14.99),
(10, 3, 1, 49.99);

-- Create a view for testing
CREATE OR REPLACE VIEW user_order_summary AS
SELECT
    u.id AS user_id,
    u.username,
    u.email,
    COUNT(o.id) AS total_orders,
    COALESCE(SUM(o.total), 0) AS total_spent
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.username, u.email;

-- Create a stored procedure for testing
DELIMITER //
CREATE PROCEDURE IF NOT EXISTS get_user_orders(IN user_id_param INT)
BEGIN
    SELECT
        o.id AS order_id,
        o.status,
        o.total,
        o.created_at,
        GROUP_CONCAT(CONCAT(p.name, ' x', oi.quantity) SEPARATOR ', ') AS items
    FROM orders o
    JOIN order_items oi ON o.id = oi.order_id
    JOIN products p ON oi.product_id = p.id
    WHERE o.user_id = user_id_param
    GROUP BY o.id, o.status, o.total, o.created_at
    ORDER BY o.created_at DESC;
END //
DELIMITER ;
