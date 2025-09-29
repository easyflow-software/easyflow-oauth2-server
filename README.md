# Easyflow OAuth2 Server
This is a own implementation of an OAuth2 server for easyflow. It follows the [RFC 6749 specification](https://datatracker.ietf.org/doc/html/rfc6749) with early adaptation for [OAuth2.1](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1-13#section-3.2).

## Setup

### Prequisites
You need to have go and pipx installed on your system for installing the needed tools.

You also need docker to run the development containers for the database.

### Setup
1. Run the `./scripts/install.sh` script to automatically install all needed tools.
2. Copy the `.env.example` file into a `.env` file and fill in missing values.
3. Start the docker containers with `docker compose up -d`.
4. You can now start the application with `./scripts/start.sh`.

## Database

### Migrations
You can create new migrations with the installed migrate tool:
```shell
migrate create -ext sql -dir pkg/database/sql/migrations -seq <migration-name>
```

To apply all migrations from the command line you can use these commands:
```shell
source .env
migrate -path $MIGRATIONS_PATH -database $DATABASE_URL up
```

To reset your db if it is in healthy condition you can execute these commands:
```shell
source .env
migrate -path $MIGRATIONS_PATH -database $DATABASE_URL down -all
```

### sqlc
For easier use of sql queries we use sqlc. More information about what sqlc is can you find here: [sqlc docs](https://docs.sqlc.dev/en/latest/index.html)

If you create new migrations or create/update queries you need to run this command to update the generated code:
```shell
sqlc generate
```
