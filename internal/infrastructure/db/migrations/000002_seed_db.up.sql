-- 000002_seed_data.up.sql

-- Insert sample users
INSERT INTO users (id, email, name, point_balance) VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'user1@example.com', 'Alice Smith', 1500),
    ('b1c9de00-0a1b-4c2d-8e3f-9a0b1c2d3e4f', 'user2@example.com', 'Bob Johnson', 750),
    ('c2d0ef11-1b2c-5d3e-9f0a-0b1c2d3e4f50', 'admin@example.com', 'Admin User', 10000);

-- Insert sample credit packages
INSERT INTO credit_packages (id, name, description, price, reward_points, is_active) VALUES
    ('f2a34b5c-6d7e-8a9b-0c1d-2e3f00010000', 'Bronze Bundle', 'Small credit package for beginners', 10.00, 100, true),
    ('f2a34b5c-6d7e-8a9b-0c1d-2e4000020000', 'Silver Bundle', 'Medium credit package with more points', 25.00, 300, true),
    ('f2a34b5c-6d7e-8a9b-0c1d-2e4100030000', 'Gold Bundle', 'Large credit package for power users', 50.00, 750, true),
    ('f2a34b5c-6d7e-8a9b-0c1d-2e4200040000', 'Platinum Bundle', 'Premium package with maximum rewards', 100.00, 1500, true);

-- Insert sample categories
INSERT INTO categories (id, name, description) VALUES
    ('1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000', 'Electronics', 'Gadgets and electronic devices'),
    ('1a2b3c4d-5e6f-7a8b-9c0d-1e3000020000', 'Books', 'Fiction and non-fiction books'),
    ('1a2b3c4d-5e6f-7a8b-9c0d-1e3100030000', 'Gift Cards', 'Various gift cards'),
    ('1a2b3c4d-5e6f-7a8b-9c0d-1e3200040000', 'Home Goods', 'Items for home improvement and decor');

-- Insert sample products
INSERT INTO products (id, name, description, category_id, point_cost, stock_quantity, is_active, is_in_offer_pool, image_url) VALUES
    ('a1b2c3d4-e5f6-a7b8-c9d0-e1f200010000', 'Wireless Earbuds', 'High-quality sound with noise cancellation.', '1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000', 500, 50, true, true, 'https://placehold.co/300x200/cccccc/333333?text=Earbuds'),
    ('a1b2c3d4-e5f6-a7b8-c9d0-e1f300020000', 'The Great Novel', 'A compelling story about adventure and discovery.', '1a2b3c4d-5e6f-7a8b-9c0d-1e3000020000', 200, 100, true, false, 'https://placehold.co/300x200/cccccc/333333?text=Novel'),
    ('a1b2c3d4-e5f6-a7b8-c9d0-e1f400030000', 'E-Reader', 'Lightweight device for digital reading.', '1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000', 800, 30, true, true, 'https://placehold.co/300x200/cccccc/333333?text=E-Reader'),
    ('a1b2c3d4-e5f6-a7b8-c9d0-e1f500040000', '10$ Gift Card', 'Digital gift card for online stores.', '1a2b3c4d-5e6f-7a8b-9c0d-1e3100030000', 1000, 200, true, true, 'https://placehold.co/300x200/cccccc/333333?text=GiftCard'),
    ('a1b2c3d4-e5f6-a7b8-c9d0-e1f600050000', 'Smart Home Hub', 'Central control for smart devices.', '1a2b3c4d-5e6f-7a8b-9c0d-1e2f00010000', 1200, 20, true, false, 'https://placehold.co/300x200/cccccc/333333?text=SmartHub'),
    ('a1b2c3d4-e5f6-a7b8-c9d0-e1f700060000', 'Cookbook: Italian Delights', 'Recipes for classic Italian dishes.', '1a2b3c4d-5e6f-7a8b-9c0d-1e3000020000', 300, 80, true, true, 'https://placehold.co/300x200/cccccc/333333?text=Cookbook');

-- Insert sample credit purchases
INSERT INTO credit_purchases (id, user_id, credit_package_id, amount_paid, points_awarded, purchase_date, status) VALUES
    ('1a2b3c4d-5e6f-7a8b-9c0d-1e2100010000', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'f2a34b5c-6d7e-8a9b-0c1d-2e4100030000', 50.00, 750, CURRENT_TIMESTAMP - INTERVAL '7 days', 'completed'),
    ('1a2b3c4d-5e6f-7a8b-9c0d-1e2200020000', 'b1c9de00-0a1b-4c2d-8e3f-9a0b1c2d3e4f', 'f2a34b5c-6d7e-8a9b-0c1d-2e4000020000', 25.00, 300, CURRENT_TIMESTAMP - INTERVAL '5 days', 'completed'),
    ('1a2b3c4d-5e6f-7a8b-9c0d-1e2300030000', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'f2a34b5c-6d7e-8a9b-0c1d-2e4000020000', 25.00, 300, CURRENT_TIMESTAMP - INTERVAL '3 days', 'completed');

-- Insert sample point redemptions
INSERT INTO point_redemptions (id, user_id, product_id, points_used, quantity, redemption_date, status) VALUES
    ('d1e2f3a4-b5c6-d7e8-f9a0-b1c200010000', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a1b2c3d4-e5f6-a7b8-c9d0-e1f200010000', 500, 1, CURRENT_TIMESTAMP - INTERVAL '2 days', 'completed'),
    ('d1e2f3a4-b5c6-d7e8-f9a0-b1c300020000', 'b1c9de00-0a1b-4c2d-8e3f-9a0b1c2d3e4f', 'a1b2c3d4-e5f6-a7b8-c9d0-e1f700060000', 300, 1, CURRENT_TIMESTAMP - INTERVAL '1 day', 'completed');
