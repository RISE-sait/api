-- +goose Up
-- +goose StatementBegin

-- Create the careers schema
CREATE SCHEMA IF NOT EXISTS careers;

-- Enum types for job postings
CREATE TYPE careers.employment_type AS ENUM ('full_time', 'part_time', 'contract', 'internship', 'volunteer');
CREATE TYPE careers.location_type AS ENUM ('on_site', 'remote', 'hybrid');
CREATE TYPE careers.job_status AS ENUM ('draft', 'published', 'closed', 'archived');
CREATE TYPE careers.application_status AS ENUM ('received', 'reviewing', 'interview', 'offer', 'hired', 'rejected', 'withdrawn');

-- Job postings table
CREATE TABLE careers.job_postings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    position TEXT NOT NULL,
    employment_type careers.employment_type NOT NULL DEFAULT 'full_time',
    location_type careers.location_type NOT NULL DEFAULT 'on_site',
    description TEXT NOT NULL,
    responsibilities TEXT[] NOT NULL DEFAULT '{}',
    requirements TEXT[] NOT NULL DEFAULT '{}',
    nice_to_have TEXT[] NOT NULL DEFAULT '{}',
    salary_min DECIMAL(10,2),
    salary_max DECIMAL(10,2),
    show_salary BOOLEAN NOT NULL DEFAULT false,
    status careers.job_status NOT NULL DEFAULT 'draft',
    closing_date TIMESTAMPTZ,
    created_by UUID REFERENCES users.users(id) ON DELETE SET NULL,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT salary_range_valid CHECK (salary_min IS NULL OR salary_max IS NULL OR salary_min <= salary_max)
);

-- Job applications table
CREATE TABLE careers.job_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES careers.job_postings(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT,
    resume_url TEXT NOT NULL,
    cover_letter TEXT,
    linkedin_url TEXT,
    portfolio_url TEXT,
    status careers.application_status NOT NULL DEFAULT 'received',
    internal_notes TEXT,
    rating INTEGER,
    reviewed_by UUID REFERENCES users.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT rating_range CHECK (rating IS NULL OR (rating >= 1 AND rating <= 5))
);

-- Indexes for job_postings
CREATE INDEX idx_job_postings_status ON careers.job_postings(status);
CREATE INDEX idx_job_postings_created_at ON careers.job_postings(created_at);

-- Indexes for job_applications
CREATE INDEX idx_job_applications_job_id ON careers.job_applications(job_id);
CREATE INDEX idx_job_applications_status ON careers.job_applications(status);
CREATE INDEX idx_job_applications_email ON careers.job_applications(email);
CREATE INDEX idx_job_applications_created_at ON careers.job_applications(created_at);

-- Auto-update triggers for updated_at
CREATE OR REPLACE FUNCTION careers.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_job_postings_updated_at
    BEFORE UPDATE ON careers.job_postings
    FOR EACH ROW
    EXECUTE FUNCTION careers.update_updated_at_column();

CREATE TRIGGER trigger_job_applications_updated_at
    BEFORE UPDATE ON careers.job_applications
    FOR EACH ROW
    EXECUTE FUNCTION careers.update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_job_applications_updated_at ON careers.job_applications;
DROP TRIGGER IF EXISTS trigger_job_postings_updated_at ON careers.job_postings;
DROP FUNCTION IF EXISTS careers.update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS careers.idx_job_applications_created_at;
DROP INDEX IF EXISTS careers.idx_job_applications_email;
DROP INDEX IF EXISTS careers.idx_job_applications_status;
DROP INDEX IF EXISTS careers.idx_job_applications_job_id;
DROP INDEX IF EXISTS careers.idx_job_postings_created_at;
DROP INDEX IF EXISTS careers.idx_job_postings_status;

-- Drop tables
DROP TABLE IF EXISTS careers.job_applications;
DROP TABLE IF EXISTS careers.job_postings;

-- Drop enum types
DROP TYPE IF EXISTS careers.application_status;
DROP TYPE IF EXISTS careers.job_status;
DROP TYPE IF EXISTS careers.location_type;
DROP TYPE IF EXISTS careers.employment_type;

-- Drop schema
DROP SCHEMA IF EXISTS careers CASCADE;

-- +goose StatementEnd
