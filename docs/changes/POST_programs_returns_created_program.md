# Fix: POST /programs Now Returns Created Program in Response Body

**Date:** 2025-12-01
**Issue:** POST /programs endpoint not returning created program in response body

---

## Problem

When creating a program via `POST /programs`, the endpoint was returning an empty response body (HTTP 201 with no content). This caused issues on the frontend when trying to upload a program photo immediately after creation.

### Previous Flow (Broken)
1. Frontend sends `POST /programs` with program data
2. Backend creates the program successfully (returns 201)
3. Backend returns **empty response body**
4. Frontend tries to call `uploadProgramPhoto(photoFile, programId, jwt)`
5. **Fails** because `programId` is `undefined`

### New Flow (Fixed)
1. Frontend sends `POST /programs` with program data
2. Backend creates the program successfully (returns 201)
3. Backend returns **the created program object with ID**
4. Frontend uses `response.id` to call `uploadProgramPhoto(photoFile, response.id, jwt)`
5. Photo upload succeeds

---

## Files Changed

### 1. `internal/domains/program/persistence/repository.go`

**Function:** `Create()`

**Before:**
```go
func (r *Repository) Create(c context.Context, details values.CreateProgramValues) *errLib.CommonError {
    // ...
    _, err := r.Queries.CreateProgram(c, dbPracticeParams)  // Result discarded!
    // ...
    return nil
}
```

**After:**
```go
func (r *Repository) Create(c context.Context, details values.CreateProgramValues) (values.GetProgramValues, *errLib.CommonError) {
    // ...
    dbProgram, err := r.Queries.CreateProgram(c, dbPracticeParams)  // Result captured
    // ...
    // Maps dbProgram to values.GetProgramValues and returns it
    return result, nil
}
```

**What changed:**
- Return type changed from `*errLib.CommonError` to `(values.GetProgramValues, *errLib.CommonError)`
- The SQL query already had `RETURNING *` so it was returning the created program, but the code was discarding it with `_`
- Now captures and maps the database result to a value object

---

### 2. `internal/domains/program/service.go`

**Function:** `CreateProgram()`

**Before:**
```go
func (s *Service) CreateProgram(ctx context.Context, details values.CreateProgramValues) *errLib.CommonError {
    // ...
    return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
        if err := txRepo.Create(ctx, details); err != nil {
            return err
        }
        // ... audit logging
    })
}
```

**After:**
```go
func (s *Service) CreateProgram(ctx context.Context, details values.CreateProgramValues) (values.GetProgramValues, *errLib.CommonError) {
    // ...
    var createdProgram values.GetProgramValues

    err := s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
        program, err := txRepo.Create(ctx, details)
        if err != nil {
            return err
        }
        createdProgram = program
        // ... audit logging
    })

    return createdProgram, err
}
```

**What changed:**
- Return type changed from `*errLib.CommonError` to `(values.GetProgramValues, *errLib.CommonError)`
- Captures the created program from the repository call
- Returns the program after the transaction completes

---

### 3. `internal/domains/program/handler.go`

**Function:** `CreateProgram()`

**Before:**
```go
func (h *Handler) CreateProgram(w http.ResponseWriter, r *http.Request) {
    // ... parse request

    if err = h.Service.CreateProgram(r.Context(), programCreate); err != nil {
        responseHandlers.RespondWithError(w, err)
        return
    }

    responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)  // nil = empty body!
}
```

**After:**
```go
func (h *Handler) CreateProgram(w http.ResponseWriter, r *http.Request) {
    // ... parse request

    program, err := h.Service.CreateProgram(r.Context(), programCreate)
    if err != nil {
        responseHandlers.RespondWithError(w, err)
        return
    }

    result := dto.Response{
        ID:          program.ID,
        Name:        program.ProgramDetails.Name,
        Description: program.ProgramDetails.Description,
        Type:        program.ProgramDetails.Type,
        CreatedAt:   program.CreatedAt,
        UpdatedAt:   program.UpdatedAt,
    }

    if program.ProgramDetails.Capacity != nil {
        result.Capacity = program.ProgramDetails.Capacity
    }

    if program.ProgramDetails.PhotoURL != nil {
        result.PhotoURL = program.ProgramDetails.PhotoURL
    }

    responseHandlers.RespondWithSuccess(w, result, http.StatusCreated)  // Returns program!
}
```

**What changed:**
- Captures the returned program from the service
- Maps it to a `dto.Response` object (same pattern used in `GetProgram` and `GetPrograms`)
- Returns the response DTO instead of `nil`
- Updated Swagger annotation from `map[string]interface{}` to `dto.Response`

---

## API Response Format

### Before
```
HTTP 201 Created
(empty body)
```

### After
```json
HTTP 201 Created
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Basketball Training",
  "description": "Weekly basketball practice sessions",
  "type": "practice",
  "capacity": 20,
  "created_at": "2025-12-01T10:30:00Z",
  "updated_at": "2025-12-01T10:30:00Z"
}
```

**Note:** `capacity` and `photo_url` are omitted if null (using `omitempty` in the DTO).

---

## Frontend Usage

After this change, the frontend can do:

```typescript
// Create the program
const response = await fetch('/programs', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${jwt}` },
  body: JSON.stringify({ name, description, type, capacity })
});

const program = await response.json();

// Now upload the photo using the returned ID
await uploadProgramPhoto(photoFile, program.id, jwt);
```

---

## Testing

To verify this fix works:

1. Send a POST request to `/programs` with valid program data
2. Confirm the response body contains the program object with an `id` field
3. Use that `id` to upload a photo or perform other operations

Example curl:
```bash
curl -X POST http://localhost/programs \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Program", "description": "Test", "type": "course", "capacity": 10}'
```

Expected: 201 response with JSON body containing `id`, `name`, `description`, `type`, `capacity`, `created_at`, `updated_at`.

---

## Notes

- No database changes required (the SQL already had `RETURNING *`)
- No breaking changes to the API contract (previously returned nothing, now returns data)
- The response structure matches `GET /programs/{id}` for consistency
