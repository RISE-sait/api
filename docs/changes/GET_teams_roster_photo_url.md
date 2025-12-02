# Feature: Add photo_url to Team Roster Members

**Date:** 2025-12-01
**Endpoint:** GET /teams/{teamId}

---

## Summary

Added `photo_url` field to each player in the team roster response. This allows the frontend to display player profile photos when viewing team details.

---

## Problem

When fetching a team by ID, the roster array returned player information but did not include their profile photo URL, even though the `photo_url` column already existed on the `athletic.athletes` table.

---

## Solution

Updated the data flow to select and return `photo_url` for each roster member.

---

## Files Changed

### 1. `internal/domains/team/persistence/sqlc/team_queries.sql`

**Change:** Added `a.photo_url` to the SELECT clause in `GetTeamRoster` query.

```sql
-- Before
SELECT u.id, u.email, u.country_alpha2_code,
       (u.first_name || ' ' || u.last_name)::varchar AS name,
       a.points, a.wins, a.losses, a.assists, a.rebounds, a.steals
FROM athletic.teams t
JOIN athletic.athletes a ON t.id = a.team_id
JOIN users.users u ON a.id = u.id
WHERE t.id = $1;

-- After
SELECT u.id, u.email, u.country_alpha2_code,
       (u.first_name || ' ' || u.last_name)::varchar AS name,
       a.points, a.wins, a.losses, a.assists, a.rebounds, a.steals,
       a.photo_url
FROM athletic.teams t
JOIN athletic.athletes a ON t.id = a.team_id
JOIN users.users u ON a.id = u.id
WHERE t.id = $1;
```

---

### 2. `internal/domains/team/values/team_details.go`

**Change:** Added `PhotoURL` field to `RosterMemberInfo` struct.

```go
type RosterMemberInfo struct {
    ID       uuid.UUID
    Email    string
    Country  string
    Name     string
    PhotoURL *string  // <-- Added
    Points   int32
    Wins     int32
    Losses   int32
    Assists  int32
    Rebounds int32
    Steals   int32
}
```

---

### 3. `internal/domains/team/persistence/repository.go`

**Change:** Map `dbMember.PhotoUrl` to the value object in `getRosterMembers()`.

```go
// Before
members[i] = values.RosterMemberInfo{
    ID:       dbMember.ID,
    Email:    dbMember.Email.String,
    // ...
}

// After
member := values.RosterMemberInfo{
    ID:       dbMember.ID,
    Email:    dbMember.Email.String,
    // ...
}

if dbMember.PhotoUrl.Valid {
    member.PhotoURL = &dbMember.PhotoUrl.String
}

members[i] = member
```

---

### 4. `internal/domains/team/dto/response_dto.go`

**Change:** Added `PhotoURL` field to DTO with JSON tag.

```go
type RosterMemberInfo struct {
    ID       uuid.UUID `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email,omitempty"`
    Country  string    `json:"country"`
    PhotoURL *string   `json:"photo_url,omitempty"`  // <-- Added
    Points   int32     `json:"points"`
    Wins     int32     `json:"wins"`
    Losses   int32     `json:"losses"`
    Assists  int32     `json:"assists"`
    Rebounds int32     `json:"rebounds"`
    Steals   int32     `json:"steals"`
}
```

---

### 5. `internal/domains/team/handler.go`

**Change:** Pass `PhotoURL` from value object to DTO in `GetTeamByID()`.

```go
roster[i] = dto.RosterMemberInfo{
    ID:       member.ID,
    Name:     member.Name,
    Email:    member.Email,
    Country:  member.Country,
    PhotoURL: member.PhotoURL,  // <-- Added
    Points:   member.Points,
    // ...
}
```

---

## Post-Pull Step

After pulling these changes, regenerate SQLC:

```bash
sqlc generate
```

This updates `persistence/sqlc/generated/team_queries.sql.go` to include `PhotoUrl sql.NullString` in the `GetTeamRosterRow` struct.

---

## API Response Example

### Before

```json
{
  "id": "team-uuid",
  "name": "Team Alpha",
  "roster": [
    {
      "id": "player-uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "country": "CA",
      "points": 150,
      "wins": 10,
      "losses": 5,
      "assists": 30,
      "rebounds": 45,
      "steals": 12
    }
  ]
}
```

### After

```json
{
  "id": "team-uuid",
  "name": "Team Alpha",
  "roster": [
    {
      "id": "player-uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "country": "CA",
      "photo_url": "https://storage.googleapis.com/bucket/player-photo.jpg",
      "points": 150,
      "wins": 10,
      "losses": 5,
      "assists": 30,
      "rebounds": 45,
      "steals": 12
    }
  ]
}
```

**Note:** `photo_url` is omitted from the response if the player doesn't have one (using `omitempty`).

---

## No Database Migration Required

The `photo_url` column already exists on `athletic.athletes` (added in migration `20250512170924_add_photo_column_staff_athlete.sql`). This change only updates the query to select it.
