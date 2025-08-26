# Rise Sports Management API

**Production**: https://api-461776259687.us-west2.run.app/  
**Square Service**: https://rise-square-service-461776259687.us-west2.run.app/

A sports management platform with Go and Python microservices handling team management, practice booking, events, and Square payment processing.

## Configuration Files

You need these files from Google Drive to run the API:
- `.env`, `.env.local`, `gcp-service-account.json`

Download them into the `config/` folder (same location as `configs.go` and `goose.yaml`).

## Local Development

### Prerequisites
- Docker and Docker Compose installed
- Configuration files in `config/` folder

### Starting the Services

```bash
# Start all services (shows logs)
docker compose up

# Or run in background
docker compose up -d

# Check services are running
docker ps
```

You should see output like:
```
api_go_server  | Server starting on :80
api_go_server  | postgresql://postgres:root@api_db:5432/mydatabase?sslmode=disable
```

Visit `http://localhost` - you should see "Welcome to Rise API".

### Database Setup

After starting the services, set up the database:

```bash
# Create tables and schemas
docker compose -f docker-compose.tasks.yml run --rm --build migrate up

# Add initial data
docker compose -f docker-compose.tasks.yml run --rm --build seed
```

Now everything should work. The API runs on `http://localhost` (port 80).

### Migrations

Done via `docker compose -f docker-compose.tasks.yml run --build --rm migrate <up/down/reset>`

All the migration files are stored as steps in `db/migrations` in the form of sql. When we perform migrate up, goose will look inside this folder and see what has and hasnt been applied yet. For unapplied ones, goose will run the sql file against the database.
U can also refer to the `cmd/migrate` directory to see some details. 

Migrate up and down are just standard commands, 'reset' is customized and its a combo of down and up, it will rollback all migrations, and then run migrate up to apply all sql in the folder.

### Seed

Done via `docker compose -f docker-compose.tasks.yml run --build --rm seed`

Seed the db with sample and real data. If u look into cmd/seed, u will see some files/functions tha are named with prefix 'fake', those are dummy data. Everything else should be able to be used for prod.

### SQLc

It is the ORM we use. In each domain, there are folders like persistence/sqlc something along the lines. Run `sqlc generate` in sqlc folder, and it will infer the sql code from `<domain>/persistence/queries` and generate Go code in the 'generated' folder.

### Service

This is the layer between the persistence ( database ) and handlers ( handle HTTP requests and responses ). We have a service layer cuz service can be thought as the business logic layer. It's where stuff like 'enroll jason in a program? first check if hes eligible, then check price based on his membership status, then generate a stripe payment link' happens. It easily involves multiple domains like payment, customer etc, which is why it has its own layer, instead of residing in handlers or persisence.

### di

This is just dependency injection

### utils/test_utils

This is necessary for testing on Github, it has a function that creates a Postgres container using docker, and db tests are run against this container.

### swagger

I stopped pushing the swagger json file to app and admin page upon merging to main, cuz otherwise it creates a lot of unnecessary commits imo. The approach now for if u wanna use swagger for type safety, is to just push the backend code as usual, which auto generates the swagger json file, then download the generated json file into your client side ( app and admin page ) repo.

Code like 

```
// GetStaffActivityLogs retrieves all staff activity logs based on filter criteria.
// @Tags staff_activity_logs
// @Summary Get staff activity logs
// @Description Retrieves a paginated list of staff activity logs with optional filtering
// @Param staff_id query string false "Filter by staff member ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param search_description query string false "Search term to filter activity descriptions (case-insensitive partial match)"
// @Param limit query int false "Number of records to return (default: 10)" example(10)
// @Param offset query int false "Number of records to skip for pagination (default: 0)" example(0)
// @Produce json
// @Success 200 {array} dto.StaffActivityLogResponse "List of staff activity logs retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input format"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /staffs/logs [get]
``` 

are swagger docs.

### db design

stuff like user info like dob, email etc are stored in our db as the source of truth,
even though we could use hubspot to store those info but it makes CRUD operations more difficult.

Like if u wanna register a user, but u successfully store some user data in our db but not hubspot

Thats why we use our db as the source of truth, and use hubspot as like an external CRM tool.

### Square integration

All Square checkout and webhook processing is handled by the Python
`square_service` microservice. The Go code under `internal/services/square` and
`internal/domains/payment/services/square_webhooks.go` is legacy.

## Production Deployment

### Go API Service

```bash
# Deploy main API
gcloud run deploy rise-web \
  --source . \
  --platform managed \
  --region us-west2 \
  --allow-unauthenticated \
  --project sacred-armor-452904-c0
```

### Python Square Service

```bash
# Go to square service directory
cd square_service

# Build the image
gcloud builds submit --config=cloudbuild.yaml .

# Deploy square service
gcloud run deploy rise-square-service \
  --image gcr.io/sacred-armor-452904-c0/rise-square-service \
  --platform managed \
  --region us-west2 \
  --allow-unauthenticated \
  --project sacred-armor-452904-c0
```

### Database Migrations

Migrations run automatically when you merge to `main` branch. For local testing:

```bash
docker compose -f docker-compose.tasks.yml run --rm --build migrate up
```

Migration files go in `db/migrations/` with this format:

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE example ADD COLUMN new_field VARCHAR(100);
-- +goose StatementEnd

-- +goose Down  
-- +goose StatementBegin
ALTER TABLE example DROP COLUMN IF EXISTS new_field;
-- +goose StatementEnd
```

### Environment Variables

Set environment variables in Google Cloud Run console for both services. Contact the team for the required values.

### Health Checks

- **Go API**: https://rise-web-461776259687.us-west2.run.app/health
- **Square Service**: https://rise-square-service-461776259687.us-west2.run.app/health

### Monitoring

Check logs in Google Cloud Console:

```bash
# Go API logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=rise-web" --project=sacred-armor-452904-c0

# Square service logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=rise-square-service" --project=sacred-armor-452904-c0
```