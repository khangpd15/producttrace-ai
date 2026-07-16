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
  "status" varchar DEFAULT 'CREATED', -- Updated default status
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
  "current_location_id" uuid, -- Renamed from current_location_point_id
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
  "purchase_location_id" uuid, -- Renamed from purchase_location_point_id
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
  "location_id" uuid, -- Renamed from location_point_id
  "event_type" varchar NOT NULL,
  "title" varchar,
  "description" text,
  "geo_location" geography,
  -- metadata_json dropped
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

CREATE TABLE "audit_logs" ( -- Renamed and reconstructed
  "id" uuid PRIMARY KEY,
  "user_id" uuid,
  "action" varchar(20),
  "entity" varchar(50),
  "entity_id" uuid,
  "old_data" jsonb,
  "new_data" jsonb,
  "created_at" timestamp DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "warranty_claims" ( -- New table
    "id" UUID PRIMARY KEY,
    "warranty_id" UUID NOT NULL,
    "product_item_id" UUID NOT NULL,
    "customer_name" VARCHAR(100) NOT NULL,
    "customer_phone" VARCHAR(20) NOT NULL,
    "customer_email" VARCHAR(150),
    "issue_title" VARCHAR(255) NOT NULL,
    "issue_description" TEXT,
    "status" VARCHAR(20) DEFAULT 'PENDING',
    "resolution_note" TEXT,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
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
ALTER TABLE "product_items" ADD FOREIGN KEY ("current_location_id") REFERENCES "locations" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "ownerships" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "ownerships" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "ownerships" ADD FOREIGN KEY ("purchase_location_id") REFERENCES "locations" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "warranties" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "warranties" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "events" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_items" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "events" ADD FOREIGN KEY ("batch_id") REFERENCES "batches" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "events" ADD FOREIGN KEY ("actor_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "events" ADD FOREIGN KEY ("location_id") REFERENCES "locations" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "attachments" ADD FOREIGN KEY ("event_id") REFERENCES "events" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "attachments" ADD FOREIGN KEY ("uploaded_by") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('ADMIN', 'STAFF', 'DEALER', 'CUSTOMER'));
ALTER TABLE users ADD CONSTRAINT chk_users_status CHECK (status IN ('ACTIVE', 'BANNED', 'SUSPENDED', 'PENDING'));
ALTER TABLE users ADD CONSTRAINT chk_users_email_not_empty CHECK (length(trim(email)) > 0); 
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email); 
CREATE INDEX uq_users_email_lower ON users (lower(email));

ALTER TABLE product_categories ADD CONSTRAINT chk_product_categories_not_self_parent CHECK (parent_id IS NULL OR id <> parent_id); 
CREATE INDEX IF NOT EXISTS idx_product_categories_parent_id ON product_categories(parent_id); 
CREATE INDEX IF NOT EXISTS idx_product_categories_slug ON product_categories(slug);

ALTER TABLE products ADD CONSTRAINT chk_products_status CHECK (status IN ('ACTIVE', 'DRAFT', 'DISCONTINUED'));
CREATE INDEX IF NOT EXISTS idx_products_tags_gin ON products USING GIN(tags);

ALTER TABLE product_variants ADD CONSTRAINT chk_product_variant_status CHECK (status IN ('ACTIVE', 'INACTIVE', 'DISCONTINUED'));
ALTER TABLE product_variants ADD CONSTRAINT chk_product_variant_price CHECK (price IS NULL OR price >= 0);
ALTER TABLE product_variants ADD CONSTRAINT chk_product_variant_currency CHECK (currency IN ('VND', 'USD'));
CREATE INDEX IF NOT EXISTS idx_product_variant_product_id ON product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_product_variant_price ON product_variants(price);

ALTER TABLE attributes ADD CONSTRAINT uq_attributes_category_code UNIQUE (category_id, code); 
CREATE INDEX IF NOT EXISTS idx_attributes_category_id ON attributes(category_id);
CREATE INDEX IF NOT EXISTS idx_attributes_code ON attributes(code);

ALTER TABLE attribute_values ADD CONSTRAINT uq_attribute_values_variant_attribute UNIQUE (product_variant_id, attribute_id); 
ALTER TABLE attribute_values ADD CONSTRAINT chk_attribute_values_only_one_value_type CHECK (
  (CASE WHEN value_text IS NOT NULL THEN 1 ELSE 0 END + CASE WHEN value_number IS NOT NULL THEN 1 ELSE 0 END + CASE WHEN value_boolean IS NOT NULL THEN 1 ELSE 0 END) = 1
); 
CREATE INDEX IF NOT EXISTS idx_attribute_values_product_variant_id ON attribute_values(product_variant_id);
CREATE INDEX IF NOT EXISTS idx_attribute_values_attribute_id ON attribute_values(attribute_id);

-- Batches Check Updated 
ALTER TABLE batches ADD CONSTRAINT chk_batches_status CHECK (status IN ('CREATED', 'IN_STOCK', 'SHIPPED', 'IN_TRANSIT', 'DELIVERED', 'SOLD_OUT', 'BLOCKED', 'RECALLED', 'CLOSED'));
ALTER TABLE batches ADD CONSTRAINT chk_batches_quantity CHECK (quantity >= 0); 
ALTER TABLE batches ADD CONSTRAINT chk_batches_date CHECK (expiry_date IS NULL OR manufacture_date IS NULL OR expiry_date >= manufacture_date);

ALTER TABLE locations ADD CONSTRAINT chk_locations_type CHECK (type IN ('WAREHOUSE', 'STORE', 'DEALER', 'WARRANTY_CENTER'));
ALTER TABLE locations ADD CONSTRAINT chk_locations_latitude CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90); 
ALTER TABLE locations ADD CONSTRAINT chk_locations_longitude CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180);
CREATE INDEX IF NOT EXISTS idx_locations_geo_location ON locations USING GIST(geo_location);

ALTER TABLE product_items ADD CONSTRAINT chk_product_items_status CHECK (status IN ('IN_STOCK', 'IN_TRANSIT', 'AT_DEALER', 'SOLD', 'REGISTERED', 'WARRANTY_ACTIVE', 'WARRANTY_CLAIMED', 'RETURNED', 'RECALLED', 'DAMAGED'));
ALTER TABLE batches ADD CONSTRAINT uq_batches_id_variant UNIQUE (id, variant_id);	
ALTER TABLE product_items ADD CONSTRAINT fk_product_items_batch_variant FOREIGN KEY (batch_id, variant_id) REFERENCES batches(id, variant_id); 
CREATE INDEX IF NOT EXISTS idx_product_items_variant_id ON product_items(variant_id);
CREATE INDEX IF NOT EXISTS idx_product_items_batch_id ON product_items(batch_id); 
CREATE INDEX IF NOT EXISTS idx_product_items_current_location_id ON product_items(current_location_id); 
ALTER TABLE product_items ADD CONSTRAINT chk_product_items_item_code_format CHECK (item_code ~ '^PTA-[0-9]{4}-[A-Z0-9]{8}$');
ALTER TABLE product_items ADD CONSTRAINT chk_product_items_verification_token_format CHECK (verification_token ~ '^[a-f0-9]{32}$');
CREATE INDEX IF NOT EXISTS idx_product_items_code_token ON product_items(item_code, verification_token);

-- Ownerships Check Updated
ALTER TABLE ownerships ADD CONSTRAINT chk_ownerships_status CHECK (status IN ('PENDING', 'ACTIVE', 'REVOKED', 'EXPIRED', 'REJECTED', 'TRANSFERRED'));
ALTER TABLE ownerships ADD CONSTRAINT chk_ownerships_type CHECK (ownership_type IN ('PRIMARY', 'TRANSFERRED'));
ALTER TABLE ownerships ADD CONSTRAINT chk_ownerships_date CHECK (ended_at IS NULL OR owned_at IS NULL OR ended_at >= owned_at);

ALTER TABLE warranties ADD CONSTRAINT chk_warranties_status CHECK (status IN ('INACTIVE', 'ACTIVE', 'EXPIRED', 'CLAIMED', 'RESOLVED', 'REJECTED', 'CANCELLED'));
ALTER TABLE warranties ADD CONSTRAINT chk_warranties_duration CHECK (duration_months IS NULL OR duration_months > 0);
ALTER TABLE warranties ADD CONSTRAINT chk_warranties_date CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date);

ALTER TABLE events ADD CONSTRAINT chk_events_event_type CHECK (event_type IN ('PRODUCED', 'PACKED', 'WAREHOUSE_IN', 'WAREHOUSE_OUT', 'TRANSPORTED', 'SALE', 'REGISTERED', 'WARRANTY_ACTIVE', 'WARRANTY_CLAIM', 'WARRANTY_RESOLVED', 'RETURNED', 'RECALL'));
ALTER TABLE events ADD CONSTRAINT chk_events_item_or_batch_required CHECK (product_item_id IS NOT NULL OR batch_id IS NOT NULL);
CREATE INDEX IF NOT EXISTS idx_events_geo_location ON events USING GIST(geo_location);

ALTER TABLE attachments ADD CONSTRAINT chk_attachments_file_type CHECK (file_type IN ('IMAGE', 'VIDEO', 'PDF', 'INVOICE', 'OTHER'));
ALTER TABLE attachments ADD CONSTRAINT chk_attachments_file_size CHECK (file_size IS NULL OR file_size >= 0);

CREATE INDEX IF NOT EXISTS idx_warranty_claims_warranty_id ON warranty_claims(warranty_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_product_item_id ON warranty_claims(product_item_id);