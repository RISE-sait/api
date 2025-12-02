# Bug Fix: registration_required Ignored When Creating Recurring Events

**Date:** 2025-12-02
**Affected Endpoint:** `POST /events/recurring`
**Severity:** Medium

## Summary

The `registration_required` field from the request payload was being ignored when creating recurring events. All created events would have `registration_required: false` regardless of the value sent in the request.

## Root Cause

The `generateEventsFromRecurrence` function in `internal/domains/event/service/service.go` was not accepting or passing the `registrationRequired` parameter to the individual events it created.

**Data Flow Before Fix:**
1. Request DTO correctly parsed `registration_required`
2. DTO passed it to `CreateRecurrenceValues.RegistrationRequired`
3. Service received it in `CreateEvents` via `details.RegistrationRequired`
4. `generateEventsFromRecurrence` was called **without** `registrationRequired`
5. Each generated event's `RegistrationRequired` defaulted to `false` (Go zero value)

## Changes Made

**File:** `internal/domains/event/service/service.go`

### 1. Updated function signature (line 556-565)

Added `registrationRequired bool` parameter:

```go
func generateEventsFromRecurrence(
    firstOccurrence, lastOccurrence time.Time,
    startTimeStr, endTimeStr string,
    mutater, programID, locationID, courtID, teamID uuid.UUID,
    membershipPlanIDs []uuid.UUID,
    priceID string,
    day time.Weekday,
    creditCost *int32,
    registrationRequired bool,  // <-- Added
) ([]values.CreateEventValues, *errLib.CommonError)
```

### 2. Updated EventDetails creation (line 598-612)

Added `RegistrationRequired` field to the generated events:

```go
events = append(events, values.CreateEventValues{
    CreatedBy: mutater,
    EventDetails: values.EventDetails{
        StartAt:                   start,
        EndAt:                     end,
        ProgramID:                 programID,
        LocationID:                locationID,
        CourtID:                   courtID,
        TeamID:                    teamID,
        RequiredMembershipPlanIDs: membershipPlanIDs,
        PriceID:                   priceID,
        CreditCost:                creditCost,
        RegistrationRequired:      registrationRequired,  // <-- Added
    },
})
```

### 3. Updated CreateEvents call (line 162-177)

Passed `details.RegistrationRequired` to the function:

```go
events, err := generateEventsFromRecurrence(
    details.FirstOccurrence,
    details.LastOccurrence,
    details.StartTime,
    details.EndTime,
    details.CreatedBy,
    details.ProgramID,
    details.LocationID,
    details.CourtID,
    details.TeamID,
    details.RequiredMembershipPlanIDs,
    details.PriceID,
    details.DayOfWeek,
    details.CreditCost,
    details.RegistrationRequired,  // <-- Added
)
```

### 4. Updated UpdateRecurringEvents call (line 390-405)

Passed `details.RegistrationRequired` to the function:

```go
eventsToCreate, err := generateEventsFromRecurrence(
    details.FirstOccurrence,
    details.LastOccurrence,
    details.StartTime,
    details.EndTime,
    details.UpdatedBy,
    details.ProgramID,
    details.LocationID,
    details.CourtID,
    details.TeamID,
    details.RequiredMembershipPlanIDs,
    details.PriceID,
    details.DayOfWeek,
    details.CreditCost,
    details.RegistrationRequired,  // <-- Added
)
```

## Testing

To verify the fix, create a recurring event with `registration_required: true`:

```json
POST /events/recurring
{
  "program_id": "uuid",
  "location_id": "uuid",
  "recurrence_start_at": "2023-10-05T07:00:00Z",
  "recurrence_end_at": "2023-10-30T07:00:00Z",
  "event_start_at": "09:00:00+00:00",
  "event_end_at": "10:00:00+00:00",
  "day": "MONDAY",
  "registration_required": true
}
```

**Expected:** Each created event in the response should have `registration_required: true`

## Notes

- The one-time event creation (`POST /events/one-time`) was already working correctly
- The default value for `registration_required` remains `true` when not provided (handled in DTO layer)
- This fix also applies to `PUT /events/recurring/{id}` for updating recurring events
