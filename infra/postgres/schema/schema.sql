CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY,
  "email" varchar UNIQUE NOT NULL,
  "phone" varchar UNIQUE,
  "full_name" text NOT NULL,
  "password_hash" varchar,
  "role" varchar DEFAULT 'CUSTOMER',
  "status" varchar DEFAULT 'ACTIVE',
  "avatar_url" text,
  "created_at" timestamp,
  "updated_at" timestamp,
  "is_deleted" boolean DEFAULT false
);

CREATE TABLE "product_categories" (
  "id" uuid PRIMARY KEY,
  "parent_id" uuid,
  "code" varchar UNIQUE,
  "name" varchar NOT NULL,
  "slug" varchar UNIQUE,
  "description" text,
  "icon_url" text,
  "is_active" boolean DEFAULT true,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE "products" (
  "id" uuid PRIMARY KEY,
  "category_id" uuid NOT NULL,
  "name" varchar NOT NULL,
  "slug" varchar UNIQUE,
  "description" text,
  "thumbnail_url" text,
  "tags" jsonb,
  "metadata_json" jsonb,
  "status" varchar DEFAULT 'ACTIVE',
  "created_by" uuid,
  "created_at" timestamp,
  "updated_at" timestamp,
  "is_deleted" boolean DEFAULT false
);

CREATE TABLE "product_variants" (
  "id" uuid PRIMARY KEY,
  "product_id" uuid NOT NULL,
  "sku" varchar UNIQUE NOT NULL,
  "name" varchar NOT NULL,
  "barcode" varchar UNIQUE,
  "price" decimal,
  "currency" varchar DEFAULT 'VND',
  "images_json" jsonb,
  "status" varchar DEFAULT 'ACTIVE',
  "created_at" timestamp,
  "updated_at" timestamp,
  "is_deleted" boolean DEFAULT false
);

CREATE TABLE "attributes" (
  "id" uuid PRIMARY KEY,
  "category_id" uuid NOT NULL,
  "code" varchar NOT NULL,
  "label" varchar NOT NULL,
  "created_at" timestamp,
  "is_deleted" boolean DEFAULT false
);

CREATE TABLE "attribute_values" (
  "id" uuid PRIMARY KEY,
  "product_variant_id" uuid NOT NULL,
  "attribute_id" uuid NOT NULL,
  "label" varchar NOT NULL,
  "value_text" text,
  "value_number" decimal,
  "value_boolean" boolean,
  "created_at" timestamp
);

CREATE TABLE "batches" (
  "id" uuid PRIMARY KEY,
  "variant_id" uuid NOT NULL,
  "batch_code" varchar UNIQUE NOT NULL,
  "manufacture_date" timestamp,
  "expiry_date" timestamp,
  "imported_at" timestamp,
  "manufacturer_name" varchar,
  "supplier_name" varchar,
  "origin_country" varchar,
  "production_place" text,
  "quantity" integer DEFAULT 0,
  "status" varchar DEFAULT 'ACTIVE',
  "created_by" uuid,
  "created_at" timestamp,
  "updated_at" timestamp,
  "is_deleted" boolean DEFAULT false
);

CREATE TABLE "locations" (
  "id" uuid PRIMARY KEY,
  "owner_user_id" uuid,
  "code" varchar UNIQUE,
  "name" varchar NOT NULL,
  "type" varchar DEFAULT 'STORE',
  "phone" varchar,
  "email" varchar,
  "address" text,
  "ward" varchar,
  "district" varchar,
  "city" varchar,
  "country" varchar DEFAULT 'Vietnam',
  "latitude" decimal,
  "longitude" decimal,
  "geo_location" geography,
  "opening_hours_json" jsonb,
  "is_active" boolean DEFAULT true,
  "created_at" timestamp,
  "updated_at" timestamp,
  "is_deleted" boolean DEFAULT false
);

CREATE TABLE "product_items" (
  "id" uuid PRIMARY KEY,
  "variant_id" uuid NOT NULL,
  "batch_id" uuid,
  "current_location_id" uuid,
  "item_code" varchar(20) UNIQUE NOT NULL,
  "verification_token" varchar(32) UNIQUE NOT NULL,
  "serial_number" varchar UNIQUE,
  "status" varchar DEFAULT 'IN_STOCK',
  "produced_at" timestamp,
  "packed_at" timestamp,
  "sold_at" timestamp,
  "registered_at" timestamp,
  "last_scanned_at" timestamp,
  "metadata_json" jsonb,
  "created_at" timestamp,
  "updated_at" timestamp,
  "is_deleted" boolean DEFAULT false
);

CREATE TABLE "ownerships" (
  "id" uuid PRIMARY KEY,
  "product_item_id" uuid NOT NULL,
  "owner_id" uuid NOT NULL,
  "status" varchar DEFAULT 'ACTIVE',
  "ownership_type" varchar DEFAULT 'PRIMARY',
  "owned_at" timestamp,
  "ended_at" timestamp,
  "purchase_date" timestamp,
  "purchase_location_id" uuid,
  "invoice_number" varchar,
  "invoice_url" text,
  "purchase_info_json" jsonb,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE "warranties" (
  "id" uuid PRIMARY KEY,
  "product_item_id" uuid NOT NULL,
  "owner_id" uuid,
  "warranty_code" varchar UNIQUE,
  "policy_name" varchar,
  "policy_description" text,
  "duration_months" integer,
  "status" varchar DEFAULT 'INACTIVE',
  "start_date" timestamp,
  "end_date" timestamp,
  "activated_at" timestamp,
  "invoice_number" varchar,
  "invoice_url" text,
  "note" text,
  "created_at" timestamp,
  "updated_at" timestamp
);

CREATE TABLE "events" (
  "id" uuid PRIMARY KEY,
  "product_item_id" uuid,
  "batch_id" uuid,
  "actor_id" uuid,
  "location_id" uuid,
  "event_type" varchar NOT NULL,
  "title" varchar,
  "description" text,
  "geo_location" geography,
  "metadata_json" jsonb,
  "created_at" timestamp
);

CREATE TABLE "attachments" (
  "id" uuid PRIMARY KEY,
  "event_id" uuid NOT NULL,
  "file_url" text NOT NULL,
  "file_public_id" varchar,
  "file_name" varchar,
  "file_type" varchar,
  "mime_type" varchar,
  "file_size" bigint,
  "uploaded_by" uuid,
  "created_at" timestamp
);

CREATE TABLE "auditLog" (
  "id" uuid PRIMARY KEY,
  "content" text NOT NULL,
  "type" varchar,
  "created_at" timestamp
);

ALTER TABLE "locations" ADD FOREIGN KEY ("owner_user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "product_categories" ADD FOREIGN KEY ("parent_id") REFERENCES "product_categories" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "products" ADD FOREIGN KEY ("category_id") REFERENCES "product_categories" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "products" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "product_variants" ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "attributes" ADD FOREIGN KEY ("category_id") REFERENCES "product_categories" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "attribute_values" ADD FOREIGN KEY ("product_variant_id") REFERENCES "product_variants" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "attribute_values" ADD FOREIGN KEY ("attribute_id") REFERENCES "attributes" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "batches" ADD FOREIGN KEY ("variant_id") REFERENCES "product_variants" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "batches" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "product_items" ADD FOREIGN KEY ("variant_id") REFERENCES "product_variants" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "product_items" ADD FOREIGN KEY ("current_location_point_id") REFERENCES "locations" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "ownerships" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "ownerships" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "ownerships" ADD FOREIGN KEY ("purchase_location_point_id") REFERENCES "locations" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "warranties" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "warranties" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "events" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "events" ADD FOREIGN KEY ("batch_id") REFERENCES "batches" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "events" ADD FOREIGN KEY ("actor_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "events" ADD FOREIGN KEY ("location_point_id") REFERENCES "locations" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "attachments" ADD FOREIGN KEY ("event_id") REFERENCES "events" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "attachments" ADD FOREIGN KEY ("uploaded_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;


ALTER TABLE users
ADD CONSTRAINT chk_users_role
CHECK (role IN (
  'ADMIN',    -- Quản trị hệ thống, có toàn quyền quản lý.
  'STAFF',    -- Nhân viên nội bộ, xử lý sản phẩm, batch, bảo hành, event.
  'DEALER',   -- Đại lý/cửa hàng phân phối sản phẩm.
  'CUSTOMER'  -- Khách hàng cuối, người mua/sở hữu sản phẩm.
));

ALTER TABLE users
ADD CONSTRAINT chk_users_status
CHECK (status IN (
  'ACTIVE',     -- Tài khoản đang hoạt động bình thường.
  'BANNED',     -- Tài khoản bị cấm vĩnh viễn hoặc dài hạn.
  'SUSPENDED'   -- Tài khoản bị tạm khóa, có thể mở lại sau.
));

ALTER TABLE users
ADD CONSTRAINT chk_users_email_not_empty
CHECK (length(trim(email)) > 0); -- Email không được là chuỗi rỗng hoặc toàn khoảng trắng.

CREATE INDEX IF NOT EXISTS idx_users_email
ON users(email); -- Tối ưu query lọc user theo email, ví dụ login.

CREATE INDEX uq_users_email_lower
ON users (lower(email));
-- Lý do: login/register bằng email, tránh trùng email do khác chữ hoa/thường.

-- =========================================================
-- PRODUCT CATEGORIES
-- =========================================================

ALTER TABLE product_categories
ADD CONSTRAINT chk_product_categories_not_self_parent
CHECK (parent_id IS NULL OR id <> parent_id); -- Chặn category tự trỏ chính nó làm parent.

CREATE INDEX IF NOT EXISTS idx_product_categories_parent_id
ON product_categories(parent_id); -- Tối ưu lấy danh mục con theo parent_id.

CREATE INDEX IF NOT EXISTS idx_product_categories_slug
ON product_categories(slug); -- Tối ưu tìm kiếm/sắp xếp category theo slug.

-- =========================================================
-- PRODUCTS
-- =========================================================

ALTER TABLE products
ADD CONSTRAINT chk_products_status
CHECK (status IN (
  'ACTIVE',        -- Sản phẩm đang được bán/hiển thị.
  'DRAFT',         -- Sản phẩm mới tạo nháp, chưa public.
  'DISCONTINUED'   -- Sản phẩm đã ngừng kinh doanh/ngừng sản xuất.
));

CREATE INDEX IF NOT EXISTS idx_products_tags_gin
ON products USING GIN(tags); -- Tối ưu search/filter trong mảng JSONB tags.

-- =========================================================
-- PRODUCT VARIANT
-- =========================================================

ALTER TABLE product_variants
ADD CONSTRAINT chk_product_variant_status
CHECK (status IN (
  'ACTIVE',        -- Variant đang bán/đang sử dụng.
  'INACTIVE',      -- Variant tạm ẩn/tạm ngưng bán.
  'DISCONTINUED'   -- Variant đã ngừng kinh doanh/ngừng sản xuất.
));

ALTER TABLE product_variants
ADD CONSTRAINT chk_product_variant_price
CHECK (price IS NULL OR price >= 0); -- Giá không được âm.

ALTER TABLE product_variants
ADD CONSTRAINT chk_product_variant_currency
CHECK (currency IN (
  'VND', -- Việt Nam Đồng.
  'USD'  -- Đô la Mỹ.
));

CREATE INDEX IF NOT EXISTS idx_product_variant_product_id
ON product_variants(product_id); -- Tối ưu lấy tất cả variant của một product.

CREATE INDEX IF NOT EXISTS idx_product_variant_price
ON product_variants(price); -- Tối ưu filter/sort theo giá.

-- =========================================================
-- ATTRIBUTES
-- =========================================================

ALTER TABLE attributes
ADD CONSTRAINT uq_attributes_category_code
UNIQUE (category_id, code); -- Một category không được có 2 attribute cùng code, ví dụ laptop không có 2 field RAM.

CREATE INDEX IF NOT EXISTS idx_attributes_category_id
ON attributes(category_id); -- Tối ưu lấy danh sách attribute theo category.

CREATE INDEX IF NOT EXISTS idx_attributes_code
ON attributes(code); -- Tối ưu tìm attribute theo code như ram, chip, storage.

-- =========================================================
-- ATTRIBUTE VALUES
-- =========================================================

ALTER TABLE attribute_values
ADD CONSTRAINT uq_attribute_values_variant_attribute
UNIQUE (product_variant_id, attribute_id); -- Một variant chỉ có một value cho một attribute, tránh iPhone có 2 RAM khác nhau.

ALTER TABLE attribute_values
ADD CONSTRAINT chk_attribute_values_only_one_value_type
CHECK (
  (
    CASE WHEN value_text IS NOT NULL THEN 1 ELSE 0 END +
    CASE WHEN value_number IS NOT NULL THEN 1 ELSE 0 END +
    CASE WHEN value_boolean IS NOT NULL THEN 1 ELSE 0 END
  ) = 1
); -- Mỗi attribute value chỉ được lưu một kiểu dữ liệu: text hoặc number hoặc boolean.

CREATE INDEX IF NOT EXISTS idx_attribute_values_product_variant_id
ON attribute_values(product_variant_id); -- Tối ưu lấy toàn bộ thông số của một variant.

CREATE INDEX IF NOT EXISTS idx_attribute_values_attribute_id
ON attribute_values(attribute_id); -- Tối ưu tìm các value theo một attribute, ví dụ tất cả RAM.

-- =========================================================
-- BATCHES
-- =========================================================


ALTER TABLE batches
ADD CONSTRAINT chk_batches_status
CHECK (status IN (
  'ACTIVE',    -- Lô hàng đang hoạt động, có thể nhập/xuất/bán.
  'CLOSED',    -- Lô hàng đã hoàn tất, thường là đã bán hết hoặc không dùng tiếp.
  'RECALLED',  -- Lô hàng bị thu hồi do lỗi sản xuất/lỗi kỹ thuật.
  'BLOCKED'    -- Lô hàng bị khóa tạm thời để kiểm tra, kiểm kê hoặc xử lý tranh chấp.
));

ALTER TABLE batches
ADD CONSTRAINT chk_batches_quantity
CHECK (quantity >= 0); -- Số lượng trong batch không được âm.

ALTER TABLE batches
ADD CONSTRAINT chk_batches_date
CHECK (
  expiry_date IS NULL
  OR manufacture_date IS NULL
  OR expiry_date >= manufacture_date
); -- Ngày hết hạn không được trước ngày sản xuất.

-- =========================================================
-- LOCATIONS
-- =========================================================


ALTER TABLE locations
ADD CONSTRAINT chk_locations_type
CHECK (type IN (
  'WAREHOUSE',        -- Kho lưu trữ hàng hóa.
  'STORE',            -- Cửa hàng bán trực tiếp.
  'DEALER',           -- Đại lý phân phối.
  'WARRANTY_CENTER'   -- Trung tâm bảo hành.
));

ALTER TABLE locations
ADD CONSTRAINT chk_locations_latitude
CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90); -- Latitude phải nằm trong biên độ địa lý hợp lệ.

ALTER TABLE locations
ADD CONSTRAINT chk_locations_longitude
CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180); -- Longitude phải nằm trong biên độ địa lý hợp lệ.

CREATE INDEX IF NOT EXISTS idx_locations_geo_location
ON locations USING GIST(geo_location); -- Tối ưu tìm location gần vị trí user bằng PostGIS.

-- =========================================================
-- PRODUCT ITEMS
-- =========================================================

ALTER TABLE product_items
ADD CONSTRAINT chk_product_items_status
CHECK (status IN (
  'IN_STOCK',          -- Item đang nằm trong kho.
  'IN_TRANSIT',        -- Item đang được vận chuyển.
  'AT_DEALER',         -- Item đang ở đại lý/cửa hàng.
  'SOLD',              -- Item đã được bán cho khách.
  'REGISTERED',        -- Item đã được khách đăng ký sở hữu/bảo hành.
  'WARRANTY_ACTIVE',   -- Item đang trong thời gian bảo hành.
  'WARRANTY_CLAIMED',  -- Item đang có yêu cầu bảo hành.
  'RETURNED',          -- Item đã bị trả lại.
  'RECALLED',          -- Item thuộc diện bị thu hồi.
  'DAMAGED'            -- Item bị hư hỏng.
));


ALTER TABLE batches
ADD CONSTRAINT uq_batches_id_variant
UNIQUE (id, variant_id);	-- Cho phép PostgreSQL reference cặp id + variant_id

ALTER TABLE product_items
ADD CONSTRAINT fk_product_items_batch_variant
FOREIGN KEY (batch_id, variant_id)
REFERENCES batches(id, variant_id); -- Đảm bảo item thuộc batch có cùng variant_id, tránh item iPhone gắn nhầm batch Samsung.

CREATE INDEX IF NOT EXISTS idx_product_items_variant_id
ON product_items(variant_id); -- Tối ưu lấy item theo variant.

CREATE INDEX IF NOT EXISTS idx_product_items_batch_id
ON product_items(batch_id); -- Tối ưu lấy item theo batch/lô hàng.

CREATE INDEX IF NOT EXISTS idx_product_items_current_location_point_id
ON product_items(current_location_point_id); -- Tối ưu tìm item đang ở location/kho nào.

ALTER TABLE product_items
ADD CONSTRAINT chk_product_items_item_code_format
CHECK (item_code ~ '^PTA-[0-9]{4}-[A-Z0-9]{8}$');
-- Format: PTA-YYMM-XXXXXXXX. Mã định danh sản phẩm vật lý, là thành phần chính trong QR code.

ALTER TABLE product_items
ADD CONSTRAINT chk_product_items_verification_token_format
CHECK (verification_token ~ '^[a-f0-9]{32}$');
-- Token bảo mật HMAC-SHA256 truncated 128 bits, dùng xác thực sản phẩm khi quét QR.

CREATE INDEX IF NOT EXISTS idx_product_items_code_token
ON product_items(item_code, verification_token);
-- Composite index tối ưu QR scan lookup: verify item_code + token cùng lúc.


-- =========================================================
-- OWNERSHIPS
-- =========================================================


ALTER TABLE ownerships
ADD CONSTRAINT chk_ownerships_status
CHECK (status IN (
  'ACTIVE',       -- Quyền sở hữu hiện tại còn hiệu lực.
  'TRANSFERRED',  -- Quyền sở hữu đã được chuyển sang người khác.
  'REVOKED'       -- Quyền sở hữu bị thu hồi/hủy bỏ.
));

ALTER TABLE ownerships
ADD CONSTRAINT chk_ownerships_type
CHECK (ownership_type IN (
  'PRIMARY',      -- Chủ sở hữu đầu tiên/sở hữu chính.
  'TRANSFERRED'   -- Chủ sở hữu nhận lại qua chuyển nhượng.
));

ALTER TABLE ownerships
ADD CONSTRAINT chk_ownerships_date
CHECK (
  ended_at IS NULL
  OR owned_at IS NULL
  OR ended_at >= owned_at
); -- Ngày kết thúc sở hữu không được trước ngày bắt đầu.

-- =========================================================
-- WARRANTIES
-- =========================================================
ALTER TABLE warranties
ADD CONSTRAINT chk_warranties_status
CHECK (status IN (
  'INACTIVE',   -- Bảo hành chưa được kích hoạt.
  'ACTIVE',     -- Bảo hành đang có hiệu lực.
  'EXPIRED',    -- Bảo hành đã hết hạn.
  'CLAIMED',    -- Đã có yêu cầu bảo hành.
  'RESOLVED',   -- Yêu cầu bảo hành đã xử lý xong.
  'REJECTED',   -- Yêu cầu bảo hành bị từ chối.
  'CANCELLED'   -- Bảo hành bị hủy.
));

ALTER TABLE warranties
ADD CONSTRAINT chk_warranties_duration
CHECK (duration_months IS NULL OR duration_months > 0); -- Thời hạn bảo hành phải lớn hơn 0 nếu có khai báo.

ALTER TABLE warranties
ADD CONSTRAINT chk_warranties_date
CHECK (
  end_date IS NULL
  OR start_date IS NULL
  OR end_date >= start_date
); -- Ngày hết hạn bảo hành không được trước ngày bắt đầu.

-- =========================================================
-- EVENTS
-- =========================================================

ALTER TABLE events
ADD CONSTRAINT chk_events_event_type
CHECK (event_type IN (
  'PRODUCED',          -- Sản phẩm/item/batch được sản xuất.
  'PACKED',            -- Sản phẩm được đóng gói.
  'WAREHOUSE_IN',      -- Nhập kho.
  'WAREHOUSE_OUT',     -- Xuất kho.
  'TRANSPORTED',       -- Đang vận chuyển.
  'SALE',              -- Bán hàng.
  'REGISTERED',        -- Khách đăng ký sản phẩm.
  'WARRANTY_ACTIVE',   -- Kích hoạt bảo hành.
  'WARRANTY_CLAIM',    -- Tạo yêu cầu bảo hành.
  'WARRANTY_RESOLVED', -- Xử lý xong bảo hành.
  'RETURNED',          -- Sản phẩm bị trả lại.
  'RECALL'             -- Sản phẩm/batch bị thu hồi.
));

ALTER TABLE events
ADD CONSTRAINT chk_events_item_or_batch_required
CHECK (
  product_item_id IS NOT NULL
  OR batch_id IS NOT NULL
); -- Event phải gắn với item hoặc batch, tránh event vô nghĩa không trace được.

CREATE INDEX IF NOT EXISTS idx_events_geo_location
ON events USING GIST(geo_location); -- Tối ưu tìm event theo vị trí địa lý bằng PostGIS.

-- =========================================================
-- ATTACHMENTS
-- =========================================================

ALTER TABLE attachments
ADD CONSTRAINT chk_attachments_file_type
CHECK (file_type IN (
  'IMAGE',    -- File hình ảnh.
  'VIDEO',    -- File video.
  'PDF',      -- File PDF.
  'INVOICE',  -- Hóa đơn/chứng từ mua bán.
  'OTHER'     -- Loại file khác.
));

ALTER TABLE attachments
ADD CONSTRAINT chk_attachments_file_size
CHECK (file_size IS NULL OR file_size >= 0); -- File size không được âm.
