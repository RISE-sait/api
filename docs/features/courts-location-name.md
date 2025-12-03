# Feature: Add location_name to Court API Responses

**Date:** 2025-12-02

## Summary

Updated `GET /courts` and `GET /courts/{id}` to include `location_name` in the response, eliminating the need for clients to make a separate API call to look up location names.

## Changes Made

### 1. SQL Queries

**File:** `internal/domains/court/persistence/sqlc/court_queries.sql`

Updated `GetCourtById` and `GetCourts` queries to JOIN with the locations table:

```sql
-- name: GetCourtById :one
SELECT c.id, c.location_id, c.name, l.name as location_name
FROM location.courts c
JOIN location.locations l ON c.location_id = l.id
WHERE c.id = $1;

-- name: GetCourts :many
SELECT c.id, c.location_id, c.name, l.name as location_name
FROM location.courts c
JOIN location.locations l ON c.location_id = l.id;
```

---

### 2. Values Struct

**File:** `internal/domains/court/values/court_details.go`

Added `LocationName` field to `ReadValues`:

```go
type ReadValues struct {
    ID           uuid.UUID
    LocationName string
    BaseDetails
}
```

---

### 3. Response DTO

**File:** `internal/domains/court/dto/response.go`

Added `LocationName` field to response:

```go
type ResponseDto struct {
    ID           uuid.UUID `json:"id"`
    Name         string    `json:"name"`
    LocationID   uuid.UUID `json:"location_id"`
    LocationName string    `json:"location_name"`
}
```

---

### 4. Repository

**File:** `internal/domains/court/persistence/repository.go`

Updated `Get` and `List` methods to map `LocationName`:

```go
return values.ReadValues{
    ID:           row.ID,
    LocationName: row.LocationName,
    BaseDetails:  values.BaseDetails{LocationID: row.LocationID, Name: row.Name},
}, nil
```

---

### 5. CI Workflow

**File:** `.github/workflows/test.yml`

Added court domain to SQLc generate step:

```yaml
cd internal/domains/court/persistence/sqlc && sqlc generate
cd ../../../../../
```

## Response Example

**Before:**
```json
{
  "id": "uuid",
  "name": "Court 1",
  "location_id": "uuid"
}
```

**After:**
```json
{
  "id": "uuid",
  "name": "Court 1",
  "location_id": "uuid",
  "location_name": "Main Gym"
}
```

## Affected Endpoints

| Endpoint | Change |
|----------|--------|
| `GET /courts` | Now includes `location_name` |
| `GET /courts/{id}` | Now includes `location_name` |

## Notes

- SQLc regeneration required after SQL query changes
- No breaking changes - `location_name` is an additive field
- No test updates required
