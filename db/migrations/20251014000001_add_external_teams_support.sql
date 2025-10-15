-- +goose Up
-- +goose StatementBegin

-- Add is_external column to differentiate between RISE teams and external opponent teams
ALTER TABLE athletic.teams
ADD COLUMN IF NOT EXISTS is_external BOOLEAN NOT NULL DEFAULT FALSE;

-- Add comment explaining external teams
COMMENT ON COLUMN athletic.teams.is_external IS 'TRUE for external/opponent teams (not RISE teams). External teams are shared across all coaches and do not require a coach_id. FALSE for internal RISE teams that must have a coach_id.';

-- Make coach_id nullable only for external teams
-- Internal teams (is_external = FALSE) must have a coach
ALTER TABLE athletic.teams
DROP CONSTRAINT IF EXISTS fk_coach;

ALTER TABLE athletic.teams
ADD CONSTRAINT fk_coach FOREIGN KEY (coach_id) REFERENCES staff.staff (id) ON DELETE SET NULL;

-- Add check constraint: internal teams must have a coach_id
ALTER TABLE athletic.teams
ADD CONSTRAINT chk_internal_team_has_coach
CHECK (is_external = TRUE OR (is_external = FALSE AND coach_id IS NOT NULL));

-- Update existing unique constraint on name to be case-insensitive
-- This prevents duplicates like "Lakers High School" and "lakers high school"
DROP INDEX IF EXISTS athletic.teams_name_key;

CREATE UNIQUE INDEX idx_teams_name_unique_case_insensitive
ON athletic.teams (LOWER(TRIM(name)));

COMMENT ON INDEX athletic.idx_teams_name_unique_case_insensitive IS 'Ensures team names are unique regardless of case or leading/trailing spaces. Prevents duplicates like "Lakers HS" and "lakers hs".';

-- Add index for filtering external vs internal teams
CREATE INDEX IF NOT EXISTS idx_teams_is_external
ON athletic.teams(is_external);

-- Add index for querying teams by coach (for "My Teams")
CREATE INDEX IF NOT EXISTS idx_teams_coach_id
ON athletic.teams(coach_id) WHERE coach_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS athletic.idx_teams_name_unique_case_insensitive;
DROP INDEX IF EXISTS athletic.idx_teams_is_external;
DROP INDEX IF EXISTS athletic.idx_teams_coach_id;

-- Remove check constraint
ALTER TABLE athletic.teams
DROP CONSTRAINT IF EXISTS chk_internal_team_has_coach;

-- Restore original foreign key
ALTER TABLE athletic.teams
DROP CONSTRAINT IF EXISTS fk_coach;

ALTER TABLE athletic.teams
ADD CONSTRAINT fk_coach FOREIGN KEY (coach_id) REFERENCES staff.staff (id) ON DELETE SET NULL;

-- Remove is_external column
ALTER TABLE athletic.teams
DROP COLUMN IF EXISTS is_external;

-- Restore original unique constraint on name
CREATE UNIQUE INDEX IF NOT EXISTS teams_name_key ON athletic.teams (name);

-- +goose StatementEnd
