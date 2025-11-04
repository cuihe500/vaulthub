# VaultHub Backend

A secure key management system built with Go.

## Quick Start

### Prerequisites

- Go 1.25.1+
- MySQL/MariaDB 5.7+

### Installation

```bash
# Clone the repository
git clone https://github.com/cuihe500/vaulthub.git
cd vaulthub/backend

# Install dependencies
go mod download

# Configure environment
cp configs/.env.example configs/.env
# Edit configs/.env with your database credentials
```

### Run

```bash
# Development mode
go run cmd/vaulthub/main.go serve

# Or build and run
go build -o vaulthub cmd/vaulthub/main.go
./vaulthub serve
```

The server will start on `http://localhost:8080`

### Health Check

```bash
curl http://localhost:8080/health
```

## Project Structure

```
.
├── cmd/vaulthub/           # Application entry point
├── internal/               # Private application code
│   ├── api/                # HTTP layer
│   ├── config/             # Configuration management
│   ├── database/           # Database layer
│   └── service/            # Business logic
├── pkg/                    # Public libraries
│   ├── crypto/             # Cryptography utilities
│   ├── logger/             # Logging utilities
│   └── response/           # HTTP response helpers
└── configs/                # Configuration files
```

## Configuration

Configuration can be set via:
1. YAML file: `configs/config.yaml`
2. Environment variables (takes precedence)

See `configs/.env.example` for all available options.

## Database Migration

VaultHub uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management.

### Auto Migration on Startup

When you run `./vaulthub serve`, migrations are automatically applied.

### Manual Migration Commands

```bash
# Apply all pending migrations
./vaulthub migrate up

# Rollback the last migration
./vaulthub migrate down

# Show current migration version
./vaulthub migrate version

# Apply N migrations (positive for up, negative for down)
./vaulthub migrate steps -n 2    # migrate up 2 steps
./vaulthub migrate steps -n -1   # migrate down 1 step

# Force set version (use with caution, only when dirty state occurs)
./vaulthub migrate force -v 1
```

### Creating New Migrations

Migration files must follow this naming convention:
- `{version}_{name}.up.sql` - forward migration
- `{version}_{name}.down.sql` - rollback migration

**Example:** Create a users table

```bash
# Create migration files
touch internal/database/migrations/000002_create_users.up.sql
touch internal/database/migrations/000002_create_users.down.sql
```

**000002_create_users.up.sql:**
```sql
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**000002_create_users.down.sql:**
```sql
DROP TABLE IF EXISTS users;
```

### Best Practices

1. **Always use `IF NOT EXISTS` / `IF EXISTS`** for idempotency
2. **Test both up and down migrations** before committing
3. **Never edit applied migrations** - create new ones instead
4. **Version numbers must increment** - use 6-digit format (000001, 000002, ...)
5. **Keep migrations atomic** - one logical change per migration

## Development

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Lint
golangci-lint run
```

## License

Apache 2.0 - see [LICENSE](../LICENSE) for details.

## Author

Changhe Cui - admin@thankseveryone.top