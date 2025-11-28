-- +goose Up

-- Add media_type to hero_promos to support videos
ALTER TABLE website.hero_promos
ADD COLUMN media_type VARCHAR(10) NOT NULL DEFAULT 'image';

-- Rename image_url to media_url for clarity
ALTER TABLE website.hero_promos
RENAME COLUMN image_url TO media_url;

-- Add thumbnail_url for video thumbnails (optional, used when media_type is 'video')
ALTER TABLE website.hero_promos
ADD COLUMN thumbnail_url TEXT;

-- Create promo_videos table for highlights section
CREATE TABLE website.promo_videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    video_url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    category VARCHAR(50) DEFAULT 'highlight',
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users.users(id),
    updated_by UUID REFERENCES users.users(id)
);

-- Index for efficient querying
CREATE INDEX idx_promo_videos_active_dates ON website.promo_videos (is_active, start_date, end_date, display_order);
CREATE INDEX idx_promo_videos_category ON website.promo_videos (category, is_active);

-- +goose Down
DROP INDEX IF EXISTS website.idx_promo_videos_category;
DROP INDEX IF EXISTS website.idx_promo_videos_active_dates;
DROP TABLE IF EXISTS website.promo_videos;

ALTER TABLE website.hero_promos DROP COLUMN IF EXISTS thumbnail_url;
ALTER TABLE website.hero_promos RENAME COLUMN media_url TO image_url;
ALTER TABLE website.hero_promos DROP COLUMN IF EXISTS media_type;
