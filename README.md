# Chirpy

A Twitter-like REST API written in Go. Users can sign up, log in, and post short messages ("chirps"). Built as a project for Boot.dev's "Learn HTTP Servers in Go" course.

Built with the Go standard library (`net/http`), PostgreSQL, JWT auth, and Argon2id password hashing.

## Requirements

- Go 1.25+
- PostgreSQL
- [goose](https://github.com/pressly/goose) for migrations
- [sqlc](https://sqlc.dev) (only needed if you change SQL queries)

## Setup

1. Create a `.env` file in the project root:

   ```env
   DB_URL="postgres://user:password@localhost:5432/chirpy?sslmode=disable"
   PLATFORM="dev"
   JWT_SECRET="<random secret>"
   POLKA_KEY="<polka api key>"
   ```

   Generate a secret with `openssl rand -base64 64`.

2. Run the migrations:

   ```sh
   goose -dir sql/schema postgres "$DB_URL" up
   ```

3. Start the server:

   ```sh
   go run .
   ```

   The server listens on `http://localhost:8080`.

## API

Authenticated endpoints expect `Authorization: Bearer <token>`.

| Method   | Path                    | Auth           | Description                                                               |
| -------- | ----------------------- | -------------- | ------------------------------------------------------------------------- |
| `GET`    | `/api/healthz`          | —              | Health check                                                              |
| `POST`   | `/api/users`            | —              | Create a user (`email`, `password`)                                       |
| `PUT`    | `/api/users`            | JWT            | Update the current user's email and password                              |
| `POST`   | `/api/login`            | —              | Log in; returns a JWT (1h) and a refresh token (60d)                      |
| `POST`   | `/api/refresh`          | Refresh token  | Get a new JWT                                                             |
| `POST`   | `/api/revoke`           | Refresh token  | Revoke a refresh token                                                    |
| `GET`    | `/api/chirps`           | —              | List chirps. Query params: `author_id`, `sort=asc\|desc` (default `desc`) |
| `GET`    | `/api/chirps/{chirpID}` | —              | Get one chirp                                                             |
| `POST`   | `/api/chirps`           | JWT            | Create a chirp (`body`, max 140 chars, profanity filtered)                |
| `DELETE` | `/api/chirps/{chirpID}` | JWT            | Delete your own chirp                                                     |
| `POST`   | `/api/polka/webhooks`   | `ApiKey <key>` | Polka webhook; `user.upgraded` enables Chirpy Red                         |
| `GET`    | `/admin/metrics`        | —              | File server hit count                                                     |
| `POST`   | `/admin/reset`          | —              | Reset hits and delete all users (`PLATFORM=dev` only)                     |

Static files are served under `/app/`.

## Development

```sh
go test ./...   # run tests
sqlc generate   # regenerate internal/database after editing sql/queries
```
