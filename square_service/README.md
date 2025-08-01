# Python Square Service

This microservice provides Square checkout functionality using FastAPI.

## Endpoints

- `POST /checkout/membership` — create a checkout link for a membership plan.
- `POST /checkout/program` — create a checkout link for a program registration.
- `POST /checkout/event` — create a checkout link for a single event.
- `POST /webhook` — handles Square payment webhooks and enrolls customers or marks registrations as paid.

Environment variables required:

- `SQUARE_ACCESS_TOKEN`
- `SQUARE_LOCATION_ID`
- `DATABASE_URL`
- `SQUARE_WEBHOOK_SIGNATURE_KEY` (optional, for verifying webhooks)

All checkout endpoints require an `Authorization: Bearer <token>` header containing a JWT with the user ID.

Run locally with:

```bash
uvicorn app:app --reload
```
