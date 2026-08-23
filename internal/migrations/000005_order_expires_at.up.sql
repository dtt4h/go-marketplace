ALTER TABLE orders ADD COLUMN expires_at TIMESTAMPTZ;

UPDATE orders
SET expires_at = created_at + INTERVAL '24 hours'
WHERE status = 'pending' AND expires_at IS NULL;

ALTER TABLE orders ALTER COLUMN expires_at SET NOT NULL;
