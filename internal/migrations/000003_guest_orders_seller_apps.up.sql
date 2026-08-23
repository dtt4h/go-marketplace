-- Guest orders: user_id becomes nullable
ALTER TABLE orders ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE orders ALTER COLUMN user_id SET DEFAULT NULL;
ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS orders_user_id_fkey,
    ADD CONSTRAINT orders_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Guest checkout + delivery fields
ALTER TABLE orders
    ADD COLUMN public_token        VARCHAR(64) UNIQUE,
    ADD COLUMN customer_first_name VARCHAR(100),
    ADD COLUMN customer_last_name  VARCHAR(100),
    ADD COLUMN customer_email      VARCHAR(255),
    ADD COLUMN customer_phone      VARCHAR(20),
    ADD COLUMN delivery_method     VARCHAR(50),
    ADD COLUMN delivery_cost       NUMERIC(12, 2) NOT NULL DEFAULT 0,
    ADD COLUMN tracking_number     VARCHAR(100);

-- Seller applications
CREATE TABLE seller_applications (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_name  VARCHAR(200) NOT NULL,
    description TEXT,
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_seller_applications_user_id ON seller_applications(user_id);
CREATE INDEX idx_seller_applications_status  ON seller_applications(status);

CREATE TRIGGER trg_seller_applications_updated_at
    BEFORE UPDATE ON seller_applications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
