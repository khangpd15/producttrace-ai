ALTER TABLE batches
DROP CONSTRAINT IF EXISTS chk_batches_status;

UPDATE batches SET status = 'IN_STOCK' WHERE status = 'ACTIVE';
UPDATE batches SET status = 'CLOSED' WHERE status = 'INACTIVE';

ALTER TABLE batches
ADD CONSTRAINT chk_batches_status
CHECK (status IN (
    'CREATED',      -- Lô vừa được tạo, chưa nhập kho.
    'IN_STOCK',     -- Lô đang ở trong kho.
    'SHIPPED',      -- Lô đã xuất kho.
    'IN_TRANSIT',   -- Lô đang được vận chuyển.
    'DELIVERED',    -- Lô đã giao tới đại lý/kho đích.
    'SOLD_OUT',     -- Toàn bộ sản phẩm trong lô đã bán hết.
    'BLOCKED',      -- Lô bị khóa để kiểm tra hoặc xử lý.
    'RECALLED',     -- Lô bị thu hồi.
    'CLOSED'        -- Lô kết thúc vòng đời.
));