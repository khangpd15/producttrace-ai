BEGIN;

-- =========================================================
-- PRODUCT ITEMS
-- =========================================================

ALTER TABLE product_items
RENAME COLUMN current_location_id TO current_location_point_id;

ALTER INDEX IF EXISTS idx_product_items_current_location_id
RENAME TO idx_product_items_current_location_point_id;

-- =========================================================
-- OWNERSHIPS
-- =========================================================

ALTER TABLE ownerships
RENAME COLUMN purchase_location_id TO purchase_location_point_id;

-- =========================================================
-- EVENTS
-- =========================================================

ALTER TABLE events
RENAME COLUMN location_id TO location_point_id;

COMMIT;