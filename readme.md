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
   `docker compose -f tasks.docker-compose.yml run --rm --build migrate up`, this setups the schemas and tables.

2. Now u got your db schemas and tables but still no data. So, run
   `docker compose -f tasks.docker-compose.yml run --rm --build seed`.

3. Now everything should work just fine.

## How to use the backend

The backend is currently running on `localhost:80`, but 80 is the default port for HTTP, so feel free to omit the 80,
making it `localhost`. Instead of sending requests to `localhost:80/courses`, u would send to `localhost/courses`.
The backend is currently running there just like any other applications, so dont worry bout any different setup just cuz
of Docker. It's just like your typical websites, so just send the request to `localhost` 
