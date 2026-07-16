CREATE TABLE IF NOT EXISTS warranty_claims (
    id UUID PRIMARY KEY,
    warranty_id UUID NOT NULL,
    product_item_id UUID NOT NULL,
    customer_name VARCHAR(100) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    customer_email VARCHAR(150),
    issue_title VARCHAR(255) NOT NULL,
    issue_description TEXT,
    status VARCHAR(20) DEFAULT 'PENDING',
    resolution_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for fast lookup
CREATE INDEX IF NOT EXISTS idx_warranty_claims_warranty_id ON warranty_claims(warranty_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_product_item_id ON warranty_claims(product_item_id);
