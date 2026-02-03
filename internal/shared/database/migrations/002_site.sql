-- Write your migrate up statements here

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS citext;

-- -----------------------------
-- 1) Accounts
-- -----------------------------
CREATE TABLE farmers (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email         CITEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  plan          TEXT NOT NULL DEFAULT 'free',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sites (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  farmer_id       UUID NOT NULL UNIQUE REFERENCES farmers(id) ON DELETE CASCADE,

  business_name   CITEXT NOT NULL UNIQUE,  -- pretty name
  business_key    CITEXT NOT NULL UNIQUE,  -- url key for lookup

  tagline         TEXT,
  location_text   TEXT,
  theme           JSONB NOT NULL DEFAULT '{}'::jsonb,
  is_published    BOOLEAN NOT NULL DEFAULT FALSE,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);


-- Helpful index for lookups
CREATE INDEX idx_sites_business_key ON sites(business_key);


-- -----------------------------
-- 3) Assets (images / logo / docs)
-- -----------------------------
CREATE TABLE assets (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  site_id      UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

  url          TEXT NOT NULL,          -- file URL/path
  kind         TEXT NOT NULL DEFAULT 'image', -- image, logo, doc
  alt_text     TEXT,
  meta         JSONB NOT NULL DEFAULT '{}'::jsonb,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_assets_site_id ON assets(site_id);

-- -----------------------------
-- 4) Single Page per site
-- -----------------------------
-- Since it’s only one page, we can just store it as "home".
CREATE TABLE pages (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  site_id      UUID NOT NULL UNIQUE REFERENCES sites(id) ON DELETE CASCADE,

  title        TEXT NOT NULL DEFAULT 'Home',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- -----------------------------
-- 5) Sections (builder)
-- -----------------------------
CREATE TABLE sections (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  page_id        UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,

  section_type   TEXT NOT NULL,  -- hero, products, trust, about, gallery, testimonials, faq, contact
  title          TEXT,
  subtitle       TEXT,

  content        JSONB NOT NULL DEFAULT '{}'::jsonb, -- flexible per type
  sort_order     INT NOT NULL DEFAULT 0,
  is_enabled     BOOLEAN NOT NULL DEFAULT TRUE,

  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

  -- Optional: prevent duplicates for singletons like hero/contact
  -- (keep products/testimonials/faq as single section too, but you might allow only one anyway)
  UNIQUE(page_id, section_type)
);

CREATE INDEX idx_sections_page_id_sort ON sections(page_id, sort_order);

-- -----------------------------
-- 6) Attach assets to sections (hero image, gallery images, logo, etc.)
-- -----------------------------
CREATE TABLE section_assets (
  section_id  UUID NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
  asset_id    UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  role        TEXT NOT NULL DEFAULT 'image', -- background, logo, gallery, etc.
  sort_order  INT NOT NULL DEFAULT 0,
  PRIMARY KEY (section_id, asset_id, role)
);

-- -----------------------------
-- 7) Products
-- -----------------------------
CREATE TABLE products (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  site_id          UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

  name            TEXT NOT NULL,
  description     TEXT,
  unit            TEXT,            -- kg, crate, bag, etc.
  min_order_qty   NUMERIC,
  is_available    BOOLEAN NOT NULL DEFAULT TRUE,
  seasonality     TEXT,

  sort_order      INT NOT NULL DEFAULT 0,

  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_products_site_id_sort ON products(site_id, sort_order);

CREATE TABLE product_assets (
  product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  asset_id    UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
  sort_order  INT NOT NULL DEFAULT 0,
  PRIMARY KEY (product_id, asset_id)
);

-- -----------------------------
-- 8) Testimonials (optional)
-- -----------------------------
CREATE TABLE testimonials (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  site_id        UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

  customer_name  TEXT NOT NULL,
  customer_role  TEXT,
  quote          TEXT NOT NULL,
  rating         INT CHECK (rating BETWEEN 1 AND 5),
  is_featured    BOOLEAN NOT NULL DEFAULT FALSE,
  sort_order     INT NOT NULL DEFAULT 0,

  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_testimonials_site_id_sort ON testimonials(site_id, sort_order);

-- -----------------------------
-- 9) FAQs (optional)
-- -----------------------------
CREATE TABLE faqs (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  site_id      UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

  question    TEXT NOT NULL,
  answer      TEXT NOT NULL,
  sort_order  INT NOT NULL DEFAULT 0,

  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_faqs_site_id_sort ON faqs(site_id, sort_order);

-- -----------------------------
-- 10) Contact profile (single per site)
-- -----------------------------
CREATE TABLE contact_profiles (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  site_id        UUID NOT NULL UNIQUE REFERENCES sites(id) ON DELETE CASCADE,

  phone          TEXT,
  whatsapp       TEXT,
  email          CITEXT,
  address        TEXT,
  service_area   TEXT,
  delivery_notes TEXT,

  socials        JSONB NOT NULL DEFAULT '{}'::jsonb,

  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
