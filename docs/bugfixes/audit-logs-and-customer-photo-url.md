# Audit Logs Improvements & Customer Photo URL

**Date:** 2025-12-02

## Summary

Improved staff activity audit logs to display human-readable names instead of UUIDs when deleting resources. Also added `photo_url` to customer API responses.

## Changes Made

### 1. Program Deletion Audit Log

**File:** `internal/domains/program/service.go`

**Before:**
```go
fmt.Sprintf("Deleted program with ID: %s", id)
```

**After:**
```go
// Fetch program name before deletion for audit log
program, err := s.repo.GetProgramByID(ctx, id)
if err != nil {
    return err
}
// ...
fmt.Sprintf("Deleted program '%s' (type: %s)", program.Name, program.Type)
```

**Audit log output:** `Deleted program 'Basketball Training' (type: course)`

---

### 2. Team Deletion Audit Log

**File:** `internal/domains/team/service.go`

**Before:**
```go
fmt.Sprintf("Deleted team with ID: %s", id)
```

**After:**
```go
// Fetch team before deletion for audit log and coach validation
team, err := s.repo.GetByID(ctx, id)
if err != nil {
    return err
}
// ...
fmt.Sprintf("Deleted team '%s'", team.TeamDetails.Name)
```

**Audit log output:** `Deleted team 'Warriors'`

**Note:** The team fetch was moved outside the coach-only block so it runs for all users, enabling the audit log to capture the team name. The coach ownership validation still works the same way.

---

### 3. Location Deletion Audit Log

**File:** `internal/domains/location/services/service.go`

**Before:**
```go
fmt.Sprintf("Deleted location with ID: %s", id)
```

**After:**
```go
// Fetch location before deletion for audit log
location, err := s.repo.GetLocationByID(ctx, id)
if err != nil {
    return err
}
// ...
fmt.Sprintf("Deleted location '%s' at %s", location.Name, location.Address)
```

**Audit log output:** `Deleted location 'Main Gym' at 123 Sports Ave`

---

### 4. Customer Photo URL in Response

**File:** `internal/domains/user/dto/customer/response.go`

Added `PhotoURL` field to the `Response` struct:

```go
type Response struct {
    // ... existing fields ...
    PhotoURL                     *string                `json:"photo_url,omitempty"`
    // ... existing fields ...
}
```

Updated `UserReadValueToResponse` to include the photo URL from athlete info:

```go
if customer.AthleteInfo != nil && customer.AthleteInfo.PhotoURL != nil {
    response.PhotoURL = customer.AthleteInfo.PhotoURL
}
```

**Affected endpoints:**
- `GET /customers/id/{id}`
- `GET /customers`
- `GET /customers/archived`

## Testing

No SQL regeneration or test updates required. All changes are in the service/DTO layer.
