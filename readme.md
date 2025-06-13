Production deployed on https://api-461776259687.us-west2.run.app/

## Google drive, get env files

To run the API, you need the .env, .env.local, and gcp-service-account.json
from https://drive.google.com/drive/folders/1t8biv_HS9dFArP6afGniQWkeuxIXu32p?usp=sharing

If you cant access the files, lemme know @ klintlee1@gmail.com.

Download the files into the config folder, which contains configs.go and goose.yaml. If you can't find those files, try
searching for them, and let me know if the issue persists.

## Running the server

1. You need Docker. If you don't have it, install it.

2. In the API folder that you just cloned ( google how to clone if you dk how ), run `docker compose up`, or optionally
   `docker compose up -d`.
   If it's your first time doing it, or if youre uncomfortable, I recommend using the former ( without the -d ). The -d
   essentially hides docker log.
   Having the log may ease things for ya.

3. Now your database and API should be running, and you can confirm this either via Docker Desktop or by running
   `docker ps`. If no, lemme know.

4. You should also see this

```
api_go_server  | 2025/03/13 22:58:34 postgresql://postgres:root@api_db:5432/mydatabase?sslmode=disable
api_go_server  | 2025/03/13 22:58:34 Server starting on :80
```

5. Go to `localhost`, it's optional to specify the port since we are using port 80, which is the default port. You
   should see something like "Welcome to Rise API". Congrats, it works. Else, youre cooked.
   At this stage, attempts to get data from the database will result in errors since the db setup is still incomplete.

## Complete db setup

1. Now your db is running but its got no data, no tables at all. So, run
   `docker compose -f docker-compose.tasks.yml run --rm --build migrate up`, this setups the schemas and tables.

2. Now u got your db schemas and tables but still no data. So, run
   `docker compose -f tasks.docker-compose.tasks.yml run --rm --build seed`.

3. Now everything should work just fine.

## How to use the backend

The backend is currently running on `localhost:80`, but 80 is the default port for HTTP, so feel free to omit the 80,
making it `localhost`. Instead of sending requests to `localhost:80/programs`, u would send to `localhost/programs`.
The backend is currently running there just like any other applications, so dont worry bout any different setup just cuz
of Docker. It's just like your typical websites, so just send the request to `localhost` 

A lot of the following concepts can be explained much better by LLMs, so I will just explain how they're implemented specific to this project.

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

This is just dependency injection, googleable.

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

Like if u wanna register a user, but u successfully store some user data in our db but not hubspot,
which would fuck things out.

Thats why we use our db as the source of truth, and use hubspot as like an external CRM tool.

rest is pretty self explanatory imo, ask me if u run into anything.