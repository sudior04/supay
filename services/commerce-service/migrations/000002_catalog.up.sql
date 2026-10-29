CREATE TABLE categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description text NOT NULL DEFAULT '' CHECK (length(description) <= 2000),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id uuid NOT NULL,
    category_id uuid NOT NULL REFERENCES categories(id),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 150),
    description text NOT NULL DEFAULT '' CHECK (length(description) <= 5000),
    status text NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','ACTIVE')),
    deleted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX products_merchant_idx ON products(merchant_id) WHERE deleted_at IS NULL;
CREATE INDEX products_category_idx ON products(category_id) WHERE deleted_at IS NULL;
CREATE TABLE product_variants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products(id),
    sku text NOT NULL CHECK (length(sku) BETWEEN 1 AND 80),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
    attributes jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(attributes)='object'),
    price_minor bigint NOT NULL CHECK (price_minor BETWEEN 1 AND 9007199254740991),
    currency char(3) NOT NULL DEFAULT 'VND' CHECK (currency='VND'),
    deleted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX variants_product_sku_idx ON product_variants(product_id,lower(sku)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX variants_product_attributes_idx ON product_variants(product_id,attributes) WHERE deleted_at IS NULL;
CREATE TABLE inventory (
    variant_id uuid PRIMARY KEY REFERENCES product_variants(id),
    quantity bigint NOT NULL DEFAULT 0 CHECK (quantity BETWEEN 0 AND 9007199254740991),
    reserved bigint NOT NULL DEFAULT 0 CHECK (reserved >= 0 AND reserved <= quantity),
    version bigint NOT NULL DEFAULT 0 CHECK (version >= 0),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE inventory_adjustments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id uuid NOT NULL REFERENCES product_variants(id),
    actor_id uuid NOT NULL,
    old_quantity bigint NOT NULL,
    new_quantity bigint NOT NULL,
    version bigint NOT NULL,
    reason text NOT NULL CHECK (length(reason) BETWEEN 1 AND 500),
    created_at timestamptz NOT NULL DEFAULT now()
);
