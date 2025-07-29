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

Run locally with:

```bash
uvicorn app:app --reload
```
