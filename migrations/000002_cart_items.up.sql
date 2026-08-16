-- cart_items
CREATE TABLE cart_items (
    id         BIGSERIAL  PRIMARY KEY,
    user_id    BIGINT     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT     NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INTEGER    NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, product_id)
);

CREATE INDEX idx_cart_items_user_id ON cart_items(user_id);
CREATE INDEX idx_cart_items_product_id ON cart_items(product_id);

CREATE TRIGGER trg_cart_items_updated_at BEFORE UPDATE ON cart_items FOR EACH ROW EXECUTE FUNCTION update_updated_at();
