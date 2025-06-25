DELETE FROM point_redemptions;

-- Deleting credit purchases (depends on users, credit_packages)
DELETE FROM credit_purchases;

-- Deleting products (depends on categories)
DELETE FROM products;

-- Now, clear parent tables that were referenced.
DELETE FROM categories;
DELETE FROM credit_packages;
DELETE FROM users;