-- +goose Up
-- Create schema for website content management
CREATE SCHEMA IF NOT EXISTS website;

-- Hero promos table (for rotating hero banners like Kapwa Tournament)
CREATE TABLE website.hero_promos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    subtitle VARCHAR(255),
    description TEXT,
    image_url TEXT NOT NULL,
    button_text VARCHAR(100),
    button_link TEXT,
    display_order INT NOT NULL DEFAULT 0,
    duration_seconds INT NOT NULL DEFAULT 5,
    is_active BOOLEAN NOT NULL DEFAULT true,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users.users(id),
    updated_by UUID REFERENCES users.users(id)
);

-- Feature cards table (for "Discover what Rise has to offer" section)
CREATE TABLE website.feature_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT NOT NULL,
    button_text VARCHAR(100),
    button_link TEXT,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users.users(id),
    updated_by UUID REFERENCES users.users(id)
);

-- Indexes for efficient querying
CREATE INDEX idx_hero_promos_active_dates ON website.hero_promos (is_active, start_date, end_date, display_order);
CREATE INDEX idx_feature_cards_active_dates ON website.feature_cards (is_active, start_date, end_date, display_order);

-- +goose Down
DROP TABLE IF EXISTS website.feature_cards;
DROP TABLE IF EXISTS website.hero_promos;
DROP SCHEMA IF EXISTS website;
