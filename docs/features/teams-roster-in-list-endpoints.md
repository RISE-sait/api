# Feature: Include Rosters in Team List Endpoints

**Date:** 2025-12-02

## Summary

Updated `GET /teams` and `GET /secure/teams` to include team rosters in the response. Previously, rosters were only returned when fetching a single team by ID (`GET /teams/{id}`).

## Changes Made

### 1. Repository - List Method

**File:** `internal/domains/team/persistence/repository.go`

Added roster fetching to the `List` method:

```go
// Fetch roster for each team
roster, rosterErr := r.getRosterMembers(ctx, dbPractice.ID)
if rosterErr != nil {
    return nil, rosterErr
}
team.Roster = roster
```

---

### 2. Repository - ListByCoach Method

**File:** `internal/domains/team/persistence/repository.go`

Added roster fetching to the `ListByCoach` method:

```go
// Fetch roster for each team
roster, rosterErr := r.getRosterMembers(ctx, dbTeam.ID)
if rosterErr != nil {
    return nil, rosterErr
}
team.Roster = roster
```

---

### 3. Handler - GetTeams

**File:** `internal/domains/team/handler.go`

Added roster mapping to the response in `GetTeams`:

```go
roster := make([]dto.RosterMemberInfo, len(team.Roster))
for j, member := range team.Roster {
    roster[j] = dto.RosterMemberInfo{
        ID:       member.ID,
        Name:     member.Name,
        Email:    member.Email,
        Country:  member.Country,
        PhotoURL: member.PhotoURL,
        Points:   member.Points,
        Wins:     member.Wins,
        Losses:   member.Losses,
        Assists:  member.Assists,
        Rebounds: member.Rebounds,
        Steals:   member.Steals,
    }
}
response.Roster = &roster
```

---

### 4. Handler - GetMyTeams (Coach Branch)

**File:** `internal/domains/team/handler.go`

Added the same roster mapping to the coach branch in `GetMyTeams`.

## Affected Endpoints

| Endpoint | Roles | Change |
|----------|-------|--------|
| `GET /teams` | All | Now includes roster for each team |
| `GET /secure/teams` | Admin, SuperAdmin, IT | Now includes roster for each team |
| `GET /secure/teams` | Coach | Now includes roster for coach's teams |

## Response Example

```json
{
  "id": "uuid",
  "name": "Warriors",
  "capacity": 15,
  "coach": {
    "id": "uuid",
    "name": "John Smith",
    "email": "john@example.com"
  },
  "logo_url": "https://...",
  "is_external": false,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "roster": [
    {
      "id": "uuid",
      "name": "Player One",
      "email": "player@example.com",
      "country": "CA",
      "photo_url": "https://...",
      "points": 150,
      "wins": 10,
      "losses": 5,
      "assists": 45,
      "rebounds": 30,
      "steals": 20
    }
  ]
}
```

## Notes

- No SQL regeneration required - uses existing `getRosterMembers` query
- No test updates required
- Swagger automatically reflects the change since `Roster` was already in the DTO
