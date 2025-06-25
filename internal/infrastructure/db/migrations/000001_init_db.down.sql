DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_credit_packages_updated_at ON credit_packages;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_products_full_text;
DROP INDEX IF EXISTS idx_categories_name;
DROP INDEX IF EXISTS idx_point_redemptions_user;
DROP INDEX IF EXISTS idx_credit_purchases_user;
DROP INDEX IF EXISTS idx_products_description_search;
DROP INDEX IF EXISTS idx_products_name_search;
DROP INDEX IF EXISTS idx_products_active;
DROP INDEX IF EXISTS idx_products_offer_pool;
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_users_point_balance;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS point_redemptions;
DROP TABLE IF EXISTS credit_purchases;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS credit_packages;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";