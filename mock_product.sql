DO $$
DECLARE
    v_category_id uuid := 'e38a221f-82ed-4b2a-89a3-5c219cb8d9e1';
    v_product_id uuid := 'a1b15174-89c0-4f51-b01c-6d8194488b39';
    v_variant_id uuid := '72d4aeec-86f1-4db8-a006-23910cabb83f';
    v_batch_id uuid := 'b66e4a2c-96b5-412e-aada-dc1c97a2cb64';
    v_item_id uuid := '86208a0d-3296-4ad9-a0ca-cb960f589e47';
    v_item_code varchar := 'PTA-2026-TEST0001';
    v_token varchar := 'a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4';
BEGIN
    IF NOT EXISTS (SELECT 1 FROM product_categories WHERE id = v_category_id) THEN
        INSERT INTO product_categories (id, name, slug)
        VALUES (v_category_id, 'Test Category', 'test-category');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM products WHERE id = v_product_id) THEN
        INSERT INTO products (id, category_id, name, slug)
        VALUES (v_product_id, v_category_id, 'Test Product', 'test-product');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM product_variants WHERE id = v_variant_id) THEN
        INSERT INTO product_variants (id, product_id, sku, name)
        VALUES (v_variant_id, v_product_id, 'TEST-SKU-001', 'Test Variant 1');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM batches WHERE id = v_batch_id) THEN
        INSERT INTO batches (id, variant_id, batch_code, status, quantity)
        VALUES (v_batch_id, v_variant_id, 'BATCH-001', 'ACTIVE', 100);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM product_items WHERE item_code = v_item_code) THEN
        INSERT INTO product_items (id, variant_id, batch_id, item_code, verification_token, serial_number, status)
        VALUES (v_item_id, v_variant_id, v_batch_id, v_item_code, v_token, 'SN-TEST-0001', 'IN_STOCK');
    END IF;
END $$;
