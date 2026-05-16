# EnergyOptimizer

EnergyOptimizer is a Go-based API for managing energy documents, using PostgreSQL as the database.

## Features

- RESTful API for CRUD operations on energy documents
- PostgreSQL database integration
- Database migrations using [golang-migrate](https://github.com/golang-migrate/migrate)
- Input validation and error handling

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.24 or newer
- [PostgreSQL](https://www.postgresql.org/download/)
- [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations

### Installation

1. **Clone the repository:**

    ```sh
    git clone https://github.com/yourusername/EnergyOptimizer.git
    cd EnergyOptimizer
    ```

2. **Install dependencies:**

    ```sh
    go mod tidy
    ```

3. **Configure your database:**
    - Create a PostgreSQL database and user.
    - Update the DSN in your config or use the `--db-dsn` flag:

      ```text
      postgres://username:password@localhost:5432/database?sslmode=disable
      ```

### Running Migrations

Install `golang-migrate` and run migrations:

```sh
migrate -source file://path/to/migrations -database postgres://username:password@localhost:5432/database?sslmode=disable up
```

### Running the API Server

```sh
go run ./cmd/api
```

You can set flags for configuration, e.g.:

```sh
go run ./cmd/api --port=4000 --env=development --db-dsn="postgres://username:password@localhost:5432/database?sslmode=disable"
```

### API Endpoints

- `POST /v1/EnergyDocument` — Create a new energy document
- `GET /v1/EnergyDocument/:id` — Get an energy document by ID
- `PUT /v1/EnergyDocument/:id` — Update an energy document
- `DELETE /v1/EnergyDocument/:id` — Delete an energy document
- `GET /v1/healthcheck` — Health check endpoint

### Project Structure

cmd/api/                # Main application entrypoint
internal/data/          # Database models and queries
internal/validator/     # Input validation logic
migrations/             # SQL migration files

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](LICENSE)
