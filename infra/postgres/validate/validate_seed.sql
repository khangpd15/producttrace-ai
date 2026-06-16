SELECT 'users' AS table_name, COUNT(*) FROM users
UNION ALL
SELECT 'shops', COUNT(*) FROM shops
UNION ALL
SELECT 'batches', COUNT(*) FROM batches
UNION ALL
SELECT 'product_items', COUNT(*) FROM product_items
UNION ALL
SELECT 'ownerships', COUNT(*) FROM ownerships
UNION ALL
SELECT 'warranties', COUNT(*) FROM warranties
UNION ALL
SELECT 'events', COUNT(*) FROM events
UNION ALL
SELECT 'attachments', COUNT(*) FROM attachments;

-- FK validation
SELECT COUNT(*) AS invalid_product_items
FROM product_items pi
LEFT JOIN batches b ON pi.batch_id = b.id
WHERE b.id IS NULL;

SELECT COUNT(*) AS invalid_ownerships
FROM ownerships o
LEFT JOIN users u ON o.owner_id = u.id
WHERE u.id IS NULL;