# Handworks API

A Go-based REST API for Handworks Cleaning Services, built with Gin, PostgreSQL, and Swagger/Redoc documentation.

## Requirements

- Go 1.25+
- PostgreSQL
- Gin framework
- Clerk (OAuth/JWT) for authentication
- Install Go Air for hot reload tooling

```bash
go install github.com/air-verse/air@latest
```

---

## Installation

1. Clone the repository:

```bash
git clone <repo-url>
cd handworks-api
```

2. Install dependencies:

```bash
go mod download
```

3. Generate Swagger docs:

```bash
swag init
```

4. Run the API:

```bash
go run main.go
```

5. Run api with hot reload:

```bash
go install github.com/air-verse/air@latest
air
```

The API runs on `http://localhost:8080` by default.

---

## API Documentation

- Open your browser and visit:
  `http://localhost:8080/swagger/index.html` (served via Swagger/swaggo)

### Authorization (Bearer Token)

1. Get your JWT token from Clerk.
2. In Swagger UI, click **Authorize**.
3. Enter:

```
Bearer <YOUR_JWT_TOKEN>
```

4. Click **Authorize** to test secured endpoints.
