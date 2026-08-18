DROP TABLE IF EXISTS seller_applications;

ALTER TABLE orders
    DROP COLUMN IF EXISTS public_token,
    DROP COLUMN IF EXISTS customer_first_name,
    DROP COLUMN IF EXISTS customer_last_name,
    DROP COLUMN IF EXISTS customer_email,
    DROP COLUMN IF EXISTS customer_phone,
    DROP COLUMN IF EXISTS delivery_method,
    DROP COLUMN IF EXISTS delivery_cost,
    DROP COLUMN IF EXISTS tracking_number;

ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS orders_user_id_fkey,
    ADD CONSTRAINT orders_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE orders ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE orders ALTER COLUMN user_id DROP DEFAULT;
