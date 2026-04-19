# Dompet Santuy API

A personal finance REST API built with Go, Echo, and MySQL. Supports JWT authentication, category management, transaction tracking, and financial summaries.

## Tech Stack

- **Go 1.25** — runtime
- **Echo v4** — HTTP framework
- **MySQL** — database
- **JWT** — access + refresh token authentication (golang-jwt/jwt v5)
- **UUID** — resource identifiers (google/uuid)

## Project Structure

```
.
├── cmd/main.go                  # Entry point, server setup, route registration
├── internal/
│   ├── config/                  # Environment config loader
│   ├── database/                # MySQL connection
│   ├── domain/                  # Request/response types and domain models
│   ├── handler/                 # HTTP handlers
│   ├── middleware/              # JWT auth middleware
│   ├── repository/              # Database queries
│   ├── response/                # Standardized response helpers
│   ├── service/                 # Business logic
│   └── util/                    # JWT manager, validator
└── migrations/                  # SQL migration files
```

## Getting Started

### Prerequisites

- Go 1.25+
- MySQL 8.0+

### Environment Variables

Create a `.env` file in the project root:

```env
APP_PORT=8080
APP_ENV=development

DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=dompet_santuy

JWT_ACCESS_SECRET=your_access_secret
JWT_REFRESH_SECRET=your_refresh_secret
JWT_ACCESS_EXPIRY_MINUTES=15
JWT_REFRESH_EXPIRY_DAYS=7
```

`JWT_ACCESS_SECRET` and `JWT_REFRESH_SECRET` are required — the server will not start without them.

### Database Setup

Run migrations in order:

```bash
mysql -u root -p < migrations/001_init.sql
mysql -u root -p < migrations/002_categories_transactions.sql
```

### Run

```bash
go run ./cmd
```

## API Reference

All protected routes require the `Authorization: Bearer <access_token>` header.

### Response Format

Every response follows a consistent envelope:

```json
{
  "success": true,
  "message": "...",
  "data": {},
  "errors": [],
  "meta": {}
}
```

- `errors` is only present on validation failures (HTTP 422)
- `meta` is only present on paginated responses
- `data` is always an array (`[]`) on list endpoints, never omitted

---

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | — | Register a new user |
| POST | `/api/v1/auth/login` | — | Login and receive tokens |
| POST | `/api/v1/auth/refresh` | — | Refresh access token |
| POST | `/api/v1/auth/logout` | — | Revoke refresh token |
| POST | `/api/v1/auth/logout-all` | ✓ | Revoke all refresh tokens |

**Register**
```json
{ "name": "Nadia Santuy", "email": "nadia@example.com", "password": "secret123" }
```

**Login / Refresh response**
```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "..."
  }
}
```

---

### User

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/me` | ✓ | Get current user profile |
| GET | `/api/v1/health` | — | Health check |

---

### Categories

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/categories` | ✓ | List all categories (filter: `?type=income\|expense`) |
| POST | `/api/v1/categories` | ✓ | Create a category |
| GET | `/api/v1/categories/:id` | ✓ | Get category by ID |
| PUT | `/api/v1/categories/:id` | ✓ | Update category |
| DELETE | `/api/v1/categories/:id` | ✓ | Delete category |

**Request body (create / update)**
```json
{
  "name": "Makan",
  "icon": "utensils",
  "color": "#F4A261",
  "type": "expense"
}
```

- `name` — required, max 100 characters
- `type` — required, `income` or `expense`
- `icon`, `color` — optional

> Deleting a category that has linked transactions returns `409 Conflict`.

---

### Transactions

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/transactions` | ✓ | List transactions (paginated) |
| POST | `/api/v1/transactions` | ✓ | Create a transaction |
| GET | `/api/v1/transactions/summary` | ✓ | Get financial summary |
| GET | `/api/v1/transactions/:id` | ✓ | Get transaction by ID |
| PUT | `/api/v1/transactions/:id` | ✓ | Update transaction |
| DELETE | `/api/v1/transactions/:id` | ✓ | Delete transaction |

**List query params**

| Param | Type | Description |
|-------|------|-------------|
| `type` | string | Filter by `income` or `expense` |
| `start_date` | string | Start date (`YYYY-MM-DD`) |
| `end_date` | string | End date (`YYYY-MM-DD`) |
| `limit` | int | Page size (default: 20, max: 100) |
| `offset` | int | Offset (default: 0) |

**Request body (create / update)**
```json
{
  "category_id": "uuid",
  "amount": 45000,
  "type": "expense",
  "note": "Ayam geprek",
  "date": "2026-04-19"
}
```

- `category_id`, `amount`, `type`, `date` — required
- `amount` must be > 0; `date` must not be in the future
- `type` must match the linked category's type

**Summary response (`GET /transactions/summary`)**

Supports optional `?start_date=&end_date=` filters.

```json
{
  "data": {
    "balance": 4955000,
    "total_income": 5000000,
    "total_expense": 45000,
    "total_count": 2,
    "income_count": 1,
    "expense_count": 1,
    "avg_income": 5000000,
    "avg_expense": 45000,
    "top_income_categories": [
      { "id": "...", "name": "Gaji", "icon": "...", "color": "...", "count": 1, "total": 5000000 }
    ],
    "top_expense_categories": [
      { "id": "...", "name": "Makan", "icon": "...", "color": "...", "count": 1, "total": 45000 }
    ]
  }
}
```

---

## Error Responses

| Status | Meaning |
|--------|---------|
| 400 | Invalid request body |
| 401 | Missing or invalid token |
| 404 | Resource not found or not owned by user |
| 409 | Conflict (e.g. email taken, category in use) |
| 422 | Validation failed — see `errors` array |
| 500 | Internal server error |

**Validation error example**
```json
{
  "success": false,
  "message": "validation failed",
  "errors": ["amount must be greater than 0", "date must not be in the future"]
}
```
